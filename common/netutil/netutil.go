package netutil

import (
	"crypto/rc4"
	"io"
	"log"
	"net"
)

func Copy(src, dst net.Conn, key string) {
	defer dst.Close()
	defer src.Close()
	var cipher *rc4.Cipher
	var err error
	if len(key) > 0 {
		cipher, err = rc4.NewCipher([]byte(key))
		if err != nil {
			log.Fatalln(err)
		}
	}
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
