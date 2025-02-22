package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/panta/machineid"
)

// Хеширует строку с использованием SHA-256 и возвращает укороченную хекс-сумму (иначе 259 символов выходит)
func hashString(data string, length int) string {
	hashSum := sha256.Sum256([]byte(data))
	fullHash := hex.EncodeToString(hashSum[:])
	if length > len(fullHash) {
		return fullHash
	}
	return fullHash[:length]
}

// Генерирует лицензионный ключ (SHA-256) и возвращает его в шестнадцатеричном виде.
func GenerateHexLicenseKey() (string, error) {
	machineID, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var macAddress string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
			macAddress = iface.HardwareAddr.String()
			break
		}
	}
	if macAddress == "" {
		return "", fmt.Errorf("no active network interface found")
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get IP addresses: %v", err)
	}

	// Ищем первый ненулевой IPv4-адрес
	var ipAddress string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ipAddress = ipNet.IP.String()
			break
		}
	}
	if ipAddress == "" {
		return "", fmt.Errorf("no active IPv4 address found")
	}

	// 8 символов для каждого параметра
	machineIDHash := hashString(machineID, 8)
	macAddressHash := hashString(macAddress, 8)
	ipAddressHash := hashString(ipAddress, 8)
	hostnameHash := hashString(hostname, 8)

	licenseKey := fmt.Sprintf("%s-%s-%s-%s", machineIDHash, macAddressHash, ipAddressHash, hostnameHash)

	return licenseKey, nil
}
