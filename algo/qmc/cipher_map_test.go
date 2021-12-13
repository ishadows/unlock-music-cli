package qmc

import (
	"os"
	"reflect"
	"testing"
)

func loadTestMapCipherData() ([]byte, []byte, []byte, error) {
	key, err := os.ReadFile("./testdata/mflac_map_key.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	raw, err := os.ReadFile("./testdata/mflac_map_raw.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	target, err := os.ReadFile("./testdata/mflac_map_target.bin")
	if err != nil {
		return nil, nil, nil, err
	}
	return key, raw, target, nil
}
func Test_mapCipher_Decrypt(t *testing.T) {
	key, raw, target, err := loadTestMapCipherData()
	if err != nil {
		t.Fatalf("load testing data failed: %s", err)
	}
	t.Run("overall", func(t *testing.T) {
		c, err := NewMapCipher(key)
		if err != nil {
			t.Errorf("init mapCipher failed: %s", err)
			return
		}
		c.Decrypt(raw, 0)
		if !reflect.DeepEqual(raw, target) {
			t.Error("overall")
		}
	})
}
