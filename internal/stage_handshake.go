package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"path"
	"time"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testHandshake(stageHarness *tester_utils.StageHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	port, err := findFreePort()
	if err != nil {
		logger.Errorf("Couldn't find free port", err)
		return err
	}
	address := fmt.Sprintf("127.0.0.1:%d", port)
	trackerURL := fmt.Sprintf("http://%s", address)
	pieceLengthBytes := 1024 * 256
	fileLengthBytes := pieceLengthBytes * len(samplePieceHashes)
	torrent := TorrentFile{
		Announce: trackerURL,
		Info: TorrentFileInfo{
			Name:        "fakefilename.iso",
			Length:      fileLengthBytes,
			Pieces:      toPiecesStr(samplePieceHashes),
			PieceLength: pieceLengthBytes,
		},
	}

	torrentFilename := "test.torrent"
	torrentOutputPath := path.Join(tempDir, torrentFilename)
	infoHash, err := torrent.writeToFile(torrentOutputPath)
	if err != nil {
		logger.Errorf("Error writing torrent file", err)
		return err
	}

	expectedPeerID, err := randomHash()
	if err != nil {
		return err
	}

	go waitAndHandlePeerConnection(address, expectedPeerID, infoHash, logger)

	logger.Debugf("Running ./your_bittorrent.sh handshake %s %s", torrentFilename, address)
	result, err := executable.Run("handshake", torrentFilename, address)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := fmt.Sprintf("Peer ID: %x\n", expectedPeerID)

	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}

func randomHash() ([20]byte, error) {
	rand.Seed(time.Now().UnixNano())
	var hash [20]byte
	if _, err := rand.Read(hash[:]); err != nil {
		return [20]byte{}, err
	}
	return hash, nil
}

func waitAndHandlePeerConnection(address string, myPeerID [20]byte, infoHash [20]byte, logger *tester_utils.Logger) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("Error:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Error accepting connection:", err)
		}
		logger.Debugf("Waiting for handshake message")
		handleConnection(conn, myPeerID, infoHash, logger)
	}
}

func handleConnection(conn net.Conn, myPeerID [20]byte, infoHash [20]byte, logger *tester_utils.Logger) {
	defer conn.Close()

	handshake, err := readHandshake(conn)
	if err != nil {
		logger.Errorf("error reading handshake", err)
		return
	}
	if !bytes.Equal(handshake.InfoHash[:], infoHash[:]) {
		logger.Errorf("expected infohash %x but got %x", infoHash, handshake.InfoHash)
		return
	}

	logger.Debugf("Received handshake: [infohash: %x, peer_id: %x]\n", handshake.InfoHash, handshake.PeerID)
	logger.Debugf("Sending back handshake")
	sendHandshake(conn, handshake.InfoHash, myPeerID)
}