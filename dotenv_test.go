package configure

import "testing"

func TestDotEnv(t *testing.T) {
	dotEnv := NewDotEnvFromFile("dotenv")

	dotEnv.Setup()

	if v, err := dotEnv.String("s"); v != "hello" {
		t.Errorf("hello %v 'hello' %v", v, err)
	}

	if v, err := dotEnv.String("S3_BUCKET"); v != "YOURS3BUCKET" {
		t.Errorf("YOURS3BUCKET %v 'YOURS3BUCKET' %v", v, err)
	}

	if v, err := dotEnv.String("HASH"); v != "#" {
		t.Errorf("# %v '#' %v", v, err)
	}

	if v, err := dotEnv.Int("INT"); v != 1 {
		t.Errorf("# %v '#' %v", v, err)
	}

	if _, err := dotEnv.String("this-message-does-not-exist"); err == nil {
		t.Error("hello2")
	}
}
