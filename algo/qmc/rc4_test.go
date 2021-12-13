package qmc

import (
	"os"
	"reflect"
	"testing"
)

func loadTestData() (*rc4Cipher, []byte, []byte, error) {
	key, err := os.ReadFile("./testdata/rc4_key.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	raw, err := os.ReadFile("./testdata/rc4_raw.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	target, err := os.ReadFile("./testdata/rc4_target.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	c, err := NewRC4Cipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	return c, raw, target, nil
}
func Test_rc4Cipher_Process(t *testing.T) {
	c, raw, target, err := loadTestData()
	if err != nil {
		t.Errorf("load testing data failed: %s", err)
	}
	t.Run("overall", func(t *testing.T) {
		c.Process(raw, 0)
		if !reflect.DeepEqual(raw, target) {
			t.Error("overall")
		}
	})

}

func Test_rc4Cipher_encFirstSegment(t *testing.T) {
	c, raw, target, err := loadTestData()
	if err != nil {
		t.Errorf("load testing data failed: %s", err)
	}
	t.Run("first-block(0~128)", func(t *testing.T) {
		c.Process(raw[:128], 0)
		if !reflect.DeepEqual(raw[:128], target[:128]) {
			t.Error("first-block(0~128)")
		}
	})
}

func Test_rc4Cipher_encASegment(t *testing.T) {
	c, raw, target, err := loadTestData()
	if err != nil {
		t.Errorf("load testing data failed: %s", err)
	}
	t.Run("align-block(128~5120)", func(t *testing.T) {
		c.Process(raw[128:5120], 128)
		if !reflect.DeepEqual(raw[128:5120], target[128:5120]) {
			t.Error("align-block(128~5120)")
		}
	})
	t.Run("simple-block(5120~10240)", func(t *testing.T) {
		c.Process(raw[5120:10240], 5120)
		if !reflect.DeepEqual(raw[5120:10240], target[5120:10240]) {
			t.Error("align-block(128~5120)")
		}
	})
}
