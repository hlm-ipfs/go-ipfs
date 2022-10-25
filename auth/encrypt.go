package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"io"
	"os"
)

const (
	saltSize                    = 12
	nonceSize                   = 12
	saltNonceSize               = saltSize + nonceSize
	keySize                     = 32
	versionCrypto        uint32 = 1
	versionSize                 = 4
	versionSaltSize             = versionSize + saltSize
	versionSaltNonceSize        = saltNonceSize + versionSize
	cryptoLineLength            = 1 * 1024
)

type cryptoGCM struct {
	salt   []byte
	nonce  []byte
	secret []byte
	key    []byte
	aesgcm cipher.AEAD
}

func NewNonce() []byte {
	nonce := make([]byte, nonceSize)
	rand.Read(nonce)
	return nonce
}

func NewKey(salt, password []byte) ([]byte, error) {
	key, err := scrypt.Key(password, salt, 16384, 8, 1, keySize)

	if err != nil {
		return nil, fmt.Errorf("Crypto New Key Failed! %v", err.Error())
	}

	return key, err
}

func NewGCM(salt, nonce, secret []byte) (*cryptoGCM, error) {
	key, err := NewKey(salt, secret)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	gcm := &cryptoGCM{
		salt:   salt,
		nonce:  nonce,
		secret: secret,
		key:    key,
		aesgcm: aesgcm,
	}

	return gcm, err
}

func (this *cryptoGCM) Decrypt(ciphertext []byte) ([]byte, error) {

	plaintext, err := this.aesgcm.Open(nil, this.nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (this *cryptoGCM) Encrypt(cleartext []byte) ([]byte, error) {

	var ciphertext []byte

	ciphertext = this.aesgcm.Seal(nil, this.nonce, cleartext, nil)

	return ciphertext, nil
}

func EncryptBigFile(src, dst string, secret []byte) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Encrypt Big File => Open srcFile Err:%v", err.Error())
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Encrypt Big File => Create dstFile Err:%v", err.Error())
	}
	defer w.Close()

	return encryptBigFile(r, w, secret)
}

func encryptBigFile(r, w *os.File, secret []byte) error {
	salt := NewNonce()
	nonce := NewNonce()

	gcm, err := NewGCM(salt, nonce, secret)
	if err != nil {
		return fmt.Errorf("Crypto New GCM Failed! %v", err.Error())
	}
	header_buf := versionedJoin(salt, nonce)
	_, err = w.Write(header_buf)
	if err != nil {
		return fmt.Errorf("Crypto Wite Encrypt File Header Err %v", err.Error())
	}

	buf := make([]byte, cryptoLineLength)
	for flag := false; ; {
		if flag {
			break
		}
		n, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("Crypto Read File Err %v", err.Error())
			}
			flag = true
		}
		if n <= 0 {
			break
		}
		ciphertext, err := gcm.Encrypt(buf[:n])
		if err != nil {
			return fmt.Errorf("Crypto GCM Encrypto Err %v", err.Error())
		}
		_, err = w.Write(ciphertext)
		if err != nil {
			return fmt.Errorf("Crypto Wirte Encrypt File Err %v", err.Error())
		}
	}
	return nil
}

func DecryptBigFile(src, dst string, secret []byte) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Decrypt Big File => Open srcFile Err:%v", err.Error())
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Decrpyt Big File => Create dstFile Err:%v", err.Error())
	}
	defer w.Close()

	return decryptBigFile(r, w, secret)
}

func decryptBigFile(r, w *os.File, secret []byte) error {
	header_buf := make([]byte, versionSaltNonceSize)
	_, err := r.Read(header_buf)
	if err != nil {
		if err != io.EOF {
		}
		return fmt.Errorf("Crypto Read Decrypt File Header Err %v", err.Error())
	}

	if len(header_buf) < versionSaltNonceSize {
		return fmt.Errorf("Crypto Read Decrypt File Header Short %v default(%v)", len(header_buf), versionSaltNonceSize)
	}

	version, salt, nonce, err := bigFileVersionedSplit(header_buf)
	if version != versionCrypto {
		return fmt.Errorf("Crypto Read Decrypt File Header Version Err %v default(%v)", version, versionCrypto)
	}
	if err != nil {
		return fmt.Errorf("Crypto Read Decrypt File Header Sum Err %v", err.Error())
	}

	gcm, err := NewGCM(salt, nonce, secret)
	if err != nil {
		return fmt.Errorf("Crypto Decrypt New GCM Err %v", err.Error())
	}

	buf_len := cryptoLineLength + gcm.aesgcm.Overhead()
	buf := make([]byte, buf_len)
	for flag := false; ; {
		if flag {
			break
		}
		n, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("Crypto Decrypt File Read Err %v", err.Error())
			}
			flag = true
		}
		if n <= 0 {
			break
		}
		decrypted, err := gcm.Decrypt(buf[:n])
		if err != nil {
			return fmt.Errorf("Crypto GCM Decrypt File Err %v", err.Error())
		}
		_, err = w.Write(decrypted)
		if err != nil {
			return fmt.Errorf("Crypto Wirte Decrypt File Err %v", err.Error())
		}
	}
	return nil
}

func versionedJoin(in ...[]byte) []byte {
	out := make([]byte, versionSize)
	binary.LittleEndian.PutUint32(out, versionCrypto)
	for _, args := range in {
		out = append(out, args...)
	}

	return out
}

func versionedSplit(in []byte) (version uint32, salt, nonce, ciphertext []byte, err error) {
	if len(in) < versionSaltNonceSize {
		return 0, nil, nil, nil, errors.New("Invalid byte length.")
	}

	version = binary.LittleEndian.Uint32(in[:versionSize])
	salt = in[versionSize:versionSaltSize]
	nonce = in[versionSaltSize:versionSaltNonceSize]
	ciphertext = in[versionSaltNonceSize:]

	return version, salt, nonce, ciphertext, nil
}

func bigFileVersionedSplit(in []byte) (version uint32, salt, nonce []byte, err error) {
	if len(in) < versionSaltNonceSize {
		return 0, nil, nil, errors.New("Invalid byte length.")
	}

	version = binary.LittleEndian.Uint32(in[:versionSize])
	salt = in[versionSize:versionSaltSize]
	nonce = in[versionSaltSize:versionSaltNonceSize]

	return version, salt, nonce, nil
}

// IsExistsPath check path exist
func IsExistsPath(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return os.IsExist(err)
	}
	return true
}
