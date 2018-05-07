package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var rootCmd = &cobra.Command{
	Use:   "ds-sync",
	Short: "syncs remote deluge files with sonarr",
	Run: func(cmd *cobra.Command, args []string) {
		err := runSSHCommand()
		fmt.Println("Result:", err)

		// sftpClient, err := sftp.NewClient(sshClient)
		// if err != nil {
		// 	panic(err)
		// }

		// sftpClient.Open("")
	},
}

func runSSHCommand() error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyFile(keyFile),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", deluge, config)
	if err != nil {
		return fmt.Errorf("Failed to dial: %s", err)
	}

	sess, err := sshClient.NewSession()
	defer sess.Close()
	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	stdin, err := createPipes(sess)
	if err != nil {
		return fmt.Errorf("failed to link pipes: %v", err)
	}

	err = sess.Run("ls ~/private/deluge/completed")
	if err != nil {
		return fmt.Errorf("failure running cmd: %v", err)
	}

	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failure reading stdin: %v", err)
	}

	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		if len(l) == 0 {
			continue
		}
		fmt.Println(l)
	}

	return nil
}

func createPipes(sess *ssh.Session) (io.Reader, error) {
	stdin, err := sess.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("Unable to setup stdin for session: %v", err)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	// go io.Copy(os.Stdout, stdout)

	stderr, err := sess.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)

	return stdout, nil
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

var (
	user, keyFile, deluge, sonarr string
)

func Execute() {
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "user to connect as")
	rootCmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "key file to connect with")
	rootCmd.Flags().StringVarP(&deluge, "deluge-host", "d", "", "deluge host to connect to")
	rootCmd.Flags().StringVarP(&sonarr, "sonarr-host", "s", "", "sonarr host to connect to")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
