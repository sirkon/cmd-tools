package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

func main() {
	var cli struct {
		Name        UserName    `arg:"" required:"" help:"User name, must be something like user, user-name, some-name2 and so on."`
		Description Description `short:"d" required:"" help:"Description of the user."`
		Commit      bool        `short:"c" help:"Add created user and restart the server."`
	}

	kong.Parse(&cli, kong.Name("awg-user"), kong.UsageOnError())
	if err := job(string(cli.Name), string(cli.Description), cli.Commit); err != nil {
		message.Fatal(err)
	}

	message.Info("done")
}

func job(name string, descr string, commit bool) error {
	clientPrivateKey, err := runCommand("", "awg", "genkey")
	if err != nil {
		return errors.Wrap(err, "run awg genkey")
	}

	clientPublicKey, err := runCommand(clientPrivateKey, "awg", "pubkey")
	if err != nil {
		return errors.Wrapf(err, "run awg pubkey over %s", clientPrivateKey)
	}

	if err := os.WriteFile(name+"_private.key", []byte(clientPrivateKey), 0644); err != nil {
		return errors.Wrap(err, "write awg client private key")
	}

	if err := os.WriteFile(name+"_public.key", []byte(clientPublicKey), 0644); err != nil {
		return errors.Wrap(err, "write awg client public key")
	}

	serverPublicKey, err := os.ReadFile("server_public.key")
	if err != nil {
		return errors.Wrap(err, "read server public key")
	}

	// Модернизируем конфиг сервера.
	const amneziaWGCfgPath = "/etc/amnezia/amneziawg/awg0.conf"
	awgConfigFile, err := os.ReadFile(amneziaWGCfgPath)
	if err != nil {
		return errors.Wrap(err, "read awg config")
	}

	awgConfigFile = bytes.TrimRight(awgConfigFile, "\n")
	awgConfigFile = append(awgConfigFile, '\n')

	goodIPPoint := 2
	for {
		ip := fmt.Sprintf("10.8.0.%d", goodIPPoint)
		if bytes.Contains(awgConfigFile, []byte(ip)) {
			goodIPPoint++
			continue
		}

		break
	}
	clientAddr := fmt.Sprintf("10.8.0.%d/32", goodIPPoint)

	tail := []string{
		"",
		"[Peer]",
		"# " + descr,
		"PublicKey = " + string(clientPublicKey),
		"AllowedIPs = " + clientAddr,
	}
	for _, t := range tail {
		awgConfigFile = append(awgConfigFile, []byte(t)...)
		awgConfigFile = append(awgConfigFile, '\n')
	}

	// Создаем конфиг клиента.
	const clientConfigTemplate = `[Interface]
PrivateKey = CLIENT_PRIVATE_KEY
Address = CLIENT_ADDRESS
DNS = 1.1.1.1
MTU = 1280

Jc = 4
Jmin = 40
Jmax = 70
S1 = 15
S2 = 27
H1 = 11111111
H2 = 22222222
H3 = 33333333
H4 = 44444444

[Peer]
PublicKey = SERVER_PUBLIC_KEY
Endpoint = 138.124.102.111:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25`
	replacer := strings.NewReplacer(
		"CLIENT_PRIVATE_KEY", clientPrivateKey,
		"CLIENT_ADDRESS", clientAddr,
		"SERVER_PUBLIC_KEY", strings.TrimSpace(string(serverPublicKey)),
	)
	clientCfg := replacer.Replace(clientConfigTemplate)

	fmt.Println("--- awg config")
	fmt.Println(string(awgConfigFile))

	fmt.Println("--- client config")
	fmt.Println(clientCfg)

	// Генерируем QR код клиента.
	qrCode, err := runCommand(clientCfg, "qrencode", "-t", "ansiutf8")
	if err != nil {
		return errors.Wrap(err, "generate QR code")
	}

	fmt.Println("--- client QR code.")
	fmt.Println(qrCode)

	if !commit {
		return nil
	}

	// Записываем конфиг клиента.
	if err := os.WriteFile(name+".conf", []byte(clientCfg), 0644); err != nil {
		return errors.Wrap(err, "write client config")
	}

	// Делаем бэкап текущего конфига
	currentConfig, err := os.ReadFile(amneziaWGCfgPath)
	if err != nil {
		return errors.Wrap(err, "read current config for backup")
	}

	const amneziaWGCfgPathBackup = amneziaWGCfgPath + ".backup"
	if err := os.WriteFile(amneziaWGCfgPathBackup, currentConfig, 0600); err != nil {
		return errors.Wrap(err, "write backup of current awg config")
	}

	if err := os.WriteFile(amneziaWGCfgPath, awgConfigFile, 0600); err != nil {
		message.Warning("failed to write awg config, you may want to roll it back with:")
		message.Warningf("sudo mv %s %s", amneziaWGCfgPathBackup, amneziaWGCfgPath)
		return errors.Wrap(err, "write awg config")
	}

	message.Info("config updated, restarting service...")
	if _, err := runCommand("", "systemctl", "restart", "awg-quick@awg0.service"); err != nil {
		message.Warning("failed to restart awg service, rolling back...")
		// Пытаемся откатить
		if rollbackErr := os.Rename(amneziaWGCfgPathBackup, amneziaWGCfgPath); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "rollback failed, system may be in inconsistent state")
		}

		// После отката пытаемся запустить сервис со старым конфигом
		if _, restartErr := runCommand("", "systemctl", "restart", "awg-quick@awg0.service"); restartErr != nil {
			return errors.Wrap(restartErr, "restart with old config failed")
		}

		return errors.Wrap(err, "restart awg with new config")
	}
	message.Info("done")

	return nil
}

func init() {
	if os.Geteuid() != 0 {
		message.Fatal("awg-user must be run as root")
		os.Exit(1)
	}
}
