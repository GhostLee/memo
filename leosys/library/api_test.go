package library

import (
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"
)

func TestUser_CheckAvailable(t *testing.T) {
		u := NewLibUser("220161223086", "153413")
		u.Login()
		//t.Log(u.CheckAvailable())
		t.Log(u.CheckIn())
	}
func TestUser_CheckFreeBook(t *testing.T) {
	u := NewLibUser("220161223086", "153413")
	u.Login()

	t.Log(u.FreeBook(6175, "720", "960", "2019-09-01"))
	t.Log(u.CheckIn())
}
func Test_getHmacEncrypt(t *testing.T) {
	id := uuid.NewV1()
	t.Log(id.String())		//ccc516dc-8686-11e9-97bc-ec8eb540ae2e
	t.Log(getHmacEncrypt("50b1d140-5970-11e9-e085-bfcd8b9885ce", "HEAD", time.Now().Unix() * 1000))
}
