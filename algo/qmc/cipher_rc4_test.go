package qmc

import (
	"os"
	"reflect"
	"testing"
)

func loadTestRC4CipherData() ([]byte, []byte, []byte, error) {
	key, err := os.ReadFile("./testdata/mflac0_rc4_key.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	raw, err := os.ReadFile("./testdata/mflac0_rc4_raw.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	target, err := os.ReadFile("./testdata/mflac0_rc4_target.bin")
	if err != nil {
		return nil, nil, nil, err
	}

	return key, raw, target, nil
}
func Test_rc4Cipher_Decrypt(t *testing.T) {
	key, raw, target, err := loadTestRC4CipherData()
	if err != nil {
		t.Fatalf("load testing data failed: %s", err)
	}
	t.Run("overall", func(t *testing.T) {
		c, err := NewRC4Cipher(key)
		if err != nil {
			t.Errorf("init rc4Cipher failed: %s", err)
			return
		}
		c.Decrypt(raw, 0)
		if !reflect.DeepEqual(raw, target) {
			t.Error("overall")
		}
	})

}

func Test_rc4Cipher_encFirstSegment(t *testing.T) {
	key, raw, target, err := loadTestRC4CipherData()
	if err != nil {
		t.Fatalf("load testing data failed: %s", err)
	}
	t.Run("first-block(0~128)", func(t *testing.T) {
		c, err := NewRC4Cipher(key)
		if err != nil {
			t.Errorf("init rc4Cipher failed: %s", err)
			return
		}
		c.Decrypt(raw[:128], 0)
		if !reflect.DeepEqual(raw[:128], target[:128]) {
			t.Error("first-block(0~128)")
		}
	})
}

func Test_rc4Cipher_encASegment(t *testing.T) {
	key, raw, target, err := loadTestRC4CipherData()
	if err != nil {
		t.Fatalf("load testing data failed: %s", err)
	}

	t.Run("align-block(128~5120)", func(t *testing.T) {
		c, err := NewRC4Cipher(key)
		if err != nil {
			t.Errorf("init rc4Cipher failed: %s", err)
			return
		}
		c.Decrypt(raw[128:5120], 128)
		if !reflect.DeepEqual(raw[128:5120], target[128:5120]) {
			t.Error("align-block(128~5120)")
		}
	})
	t.Run("simple-block(5120~10240)", func(t *testing.T) {
		c, err := NewRC4Cipher(key)
		if err != nil {
			t.Errorf("init rc4Cipher failed: %s", err)
			return
		}
		c.Decrypt(raw[5120:10240], 5120)
		if !reflect.DeepEqual(raw[5120:10240], target[5120:10240]) {
			t.Error("align-block(128~5120)")
		}
	})
}
