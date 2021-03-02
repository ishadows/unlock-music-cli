package ncm

import (
	"github.com/unlock-music/cli/algo/common"
	"strings"
)

type RawMeta interface {
	common.Meta
	GetFormat() string
	GetAlbumImageURL() string
}
type RawMetaMusic struct {
	Format        string          `json:"format"`
	MusicID       int             `json:"musicId"`
	MusicName     string          `json:"musicName"`
	Artist        [][]interface{} `json:"artist"`
	Album         string          `json:"album"`
	AlbumID       int             `json:"albumId"`
	AlbumPicDocID interface{}     `json:"albumPicDocId"`
	AlbumPic      string          `json:"albumPic"`
	MvID          int             `json:"mvId"`
	Flag          int             `json:"flag"`
	Bitrate       int             `json:"bitrate"`
	Duration      int             `json:"duration"`
	Alias         []interface{}   `json:"alias"`
	TransNames    []interface{}   `json:"transNames"`
}

func (m RawMetaMusic) GetAlbumImageURL() string {
	return m.AlbumPic
}
func (m RawMetaMusic) GetArtists() (artists []string) {
	for _, artist := range m.Artist {
		for _, item := range artist {
			name, ok := item.(string)
			if ok {
				artists = append(artists, name)
			}
		}
	}
	return
}

func (m RawMetaMusic) GetTitle() string {
	return m.MusicName
}

func (m RawMetaMusic) GetAlbum() string {
	return m.Album
}
func (m RawMetaMusic) GetFormat() string {
	return m.Format
}

//goland:noinspection SpellCheckingInspection
type RawMetaDJ struct {
	ProgramID          int          `json:"programId"`
	ProgramName        string       `json:"programName"`
	MainMusic          RawMetaMusic `json:"mainMusic"`
	DjID               int          `json:"djId"`
	DjName             string       `json:"djName"`
	DjAvatarURL        string       `json:"djAvatarUrl"`
	CreateTime         int64        `json:"createTime"`
	Brand              string       `json:"brand"`
	Serial             int          `json:"serial"`
	ProgramDesc        string       `json:"programDesc"`
	ProgramFeeType     int          `json:"programFeeType"`
	ProgramBuyed       bool         `json:"programBuyed"`
	RadioID            int          `json:"radioId"`
	RadioName          string       `json:"radioName"`
	RadioCategory      string       `json:"radioCategory"`
	RadioCategoryID    int          `json:"radioCategoryId"`
	RadioDesc          string       `json:"radioDesc"`
	RadioFeeType       int          `json:"radioFeeType"`
	RadioFeeScope      int          `json:"radioFeeScope"`
	RadioBuyed         bool         `json:"radioBuyed"`
	RadioPrice         int          `json:"radioPrice"`
	RadioPurchaseCount int          `json:"radioPurchaseCount"`
}

func (m RawMetaDJ) GetArtists() []string {
	if m.DjName != "" {
		return []string{m.DjName}
	}
	return m.MainMusic.GetArtists()
}

func (m RawMetaDJ) GetTitle() string {
	if m.ProgramName != "" {
		return m.ProgramName
	}
	return m.MainMusic.GetTitle()
}

func (m RawMetaDJ) GetAlbum() string {
	if m.Brand != "" {
		return m.Brand
	}
	return m.MainMusic.GetAlbum()
}

func (m RawMetaDJ) GetFormat() string {
	return m.MainMusic.GetFormat()
}

func (m RawMetaDJ) GetAlbumImageURL() string {
	if strings.HasPrefix(m.MainMusic.GetAlbumImageURL(), "http") {
		return m.MainMusic.GetAlbumImageURL()
	}
	return m.DjAvatarURL
}
