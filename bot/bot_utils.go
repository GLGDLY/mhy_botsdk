package bot

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"

	crypto4go "github.com/smartwalle/crypto4go"
)

// replace space with \n and add \n at the end
func parsePubKey(pubKey string) string {
	k := strings.Replace(pubKey, "-----BEGIN PUBLIC KEY-----", "", -1)
	k = strings.Replace(k, "-----END PUBLIC KEY-----", "", -1)
	k = strings.TrimSpace(k)

	k = strings.Replace(k, " ", "\n", -1)
	k = "-----BEGIN PUBLIC KEY-----" + "\n" + k + "\n" + "-----END PUBLIC KEY-----\n"
	return k
}

func pubKeyEncryptSecret(pubKey string, botSecret string) string {
	h := hmac.New(sha256.New, []byte(pubKey))
	raw := []byte(botSecret)
	h.Write(raw)
	return hex.EncodeToString(h.Sum(nil))
}

func pubKeyVerify(sign, body, botSecret, pubKey string) (bool, error) {
	signArg, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}
	str := url.Values{
		"body":   {strings.TrimSpace(body)},
		"secret": {botSecret},
	}.Encode()

	hashedOrigin := sha256.Sum256([]byte(str))
	pub, _ := crypto4go.ParsePublicKey(crypto4go.FormatPublicKey(pubKey))
	if err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashedOrigin[:], signArg); err != nil {
		return false, err
	}
	return true, nil
}
