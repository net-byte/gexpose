package netutil

import (
	"crypto/rc4"
	"io"
	"net"
)

func Copy(src, dst net.Conn, key string) {
	if src == nil || dst == nil || key == "" {
		return
	}
	defer dst.Close()
	defer src.Close()
	cipher, _ := rc4.NewCipher([]byte(key))
	buf := make([]byte, 64*1024)
	for {
		n, err := src.Read(buf)
		if err != nil || err == io.EOF {
			break
		}
		b := buf[:n]
		if cipher != nil {
			cipher.XORKeyStream(b, b)
		}
		_, err = dst.Write(b)
		if err != nil {
			break
		}
	}
}
