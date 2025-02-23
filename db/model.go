// Package db provides database helpers and models
//
//nolint:lll // struct tags get very long and can't be split
package db

// see this db fiddle to mess around with the schema
// https://www.db-fiddle.com/f/wJ7z8L7mu6ZKaYmWk1xr1p/5

import (
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.senan.xyz/gonic/mime"

	// TODO: remove this dep
	"go.senan.xyz/gonic/server/ctrlsubsonic/specid"
)

func splitIDs(in, sep string) []specid.ID {
	if in == "" {
		return []specid.ID{}
	}
	parts := strings.Split(in, sep)
	ret := make([]specid.ID, 0, len(parts))
	for _, p := range parts {
		id, _ := specid.New(p)
		ret = append(ret, id)
	}
	return ret
}

func joinIds(in []specid.ID, sep string) string {
	if in == nil {
		return ""
	}
	strs := make([]string, 0, len(in))
	for _, id := range in {
		strs = append(strs, id.String())
	}
	return strings.Join(strs, sep)
}

type Artist struct {
	ID            int      `gorm:"primary_key"`
	Name          string   `gorm:"not null; unique_index"`
	NameUDec      string   `sql:"default: null"`
	Albums        []*Album `gorm:"many2many:album_artists"`
	AlbumCount    int      `sql:"-"`
	ArtistStar    *ArtistStar
	ArtistRating  *ArtistRating
	AverageRating float64 `sql:"default: null"`
}

func (a *Artist) SID() *specid.ID {
	return &specid.ID{Type: specid.Artist, Value: a.ID}
}

func (a *Artist) IndexName() string {
	if len(a.NameUDec) > 0 {
		return a.NameUDec
	}
	return a.Name
}

type Genre struct {
	ID         int    `gorm:"primary_key"`
	Name       string `gorm:"not null; unique_index"`
	AlbumCount int    `sql:"-"`
	TrackCount int    `sql:"-"`
}

// AudioFile is used to avoid some duplication in handlers_raw.go
// between Track and Podcast
type AudioFile interface {
	Ext() string
	MIME() string
	AudioFilename() string
	AudioBitrate() int
	AudioLength() int
}

type Track struct {
	ID             int `gorm:"primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Filename       string `gorm:"not null; unique_index:idx_folder_filename" sql:"default: null"`
	FilenameUDec   string `sql:"default: null"`
	Album          *Album
	AlbumID        int      `gorm:"not null; unique_index:idx_folder_filename" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	Genres         []*Genre `gorm:"many2many:track_genres"`
	Size           int      `sql:"default: null"`
	Length         int      `sql:"default: null"`
	Bitrate        int      `sql:"default: null"`
	TagTitle       string   `sql:"default: null"`
	TagTitleUDec   string   `sql:"default: null"`
	TagTrackArtist string   `sql:"default: null"`
	TagTrackNumber int      `sql:"default: null"`
	TagDiscNumber  int      `sql:"default: null"`
	TagBrainzID    string   `sql:"default: null"`
	TrackStar      *TrackStar
	TrackRating    *TrackRating
	AverageRating  float64 `sql:"default: null"`
}

func (t *Track) AudioLength() int  { return t.Length }
func (t *Track) AudioBitrate() int { return t.Bitrate }

func (t *Track) SID() *specid.ID {
	return &specid.ID{Type: specid.Track, Value: t.ID}
}

func (t *Track) AlbumSID() *specid.ID {
	return &specid.ID{Type: specid.Album, Value: t.AlbumID}
}

func (t *Track) Ext() string {
	return filepath.Ext(t.Filename)
}

func (t *Track) AudioFilename() string {
	return t.Filename
}

func (t *Track) MIME() string {
	return mime.TypeByExtension(filepath.Ext(t.Filename))
}

func (t *Track) AbsPath() string {
	if t.Album == nil {
		return ""
	}
	return path.Join(
		t.Album.RootDir,
		t.Album.LeftPath,
		t.Album.RightPath,
		t.Filename,
	)
}

func (t *Track) RelPath() string {
	if t.Album == nil {
		return ""
	}
	return path.Join(
		t.Album.LeftPath,
		t.Album.RightPath,
		t.Filename,
	)
}

func (t *Track) GenreStrings() []string {
	strs := make([]string, 0, len(t.Genres))
	for _, genre := range t.Genres {
		strs = append(strs, genre.Name)
	}
	return strs
}

type User struct {
	ID                int `gorm:"primary_key"`
	CreatedAt         time.Time
	Name              string `gorm:"not null; unique_index" sql:"default: null"`
	Password          string `gorm:"not null" sql:"default: null"`
	LastFMSession     string `sql:"default: null"`
	ListenBrainzURL   string `sql:"default: null"`
	ListenBrainzToken string `sql:"default: null"`
	IsAdmin           bool   `sql:"default: null"`
	Avatar            []byte `sql:"default: null"`
}

type Setting struct {
	Key   string `gorm:"not null; primary_key; auto_increment:false" sql:"default: null"`
	Value string `sql:"default: null"`
}

type Play struct {
	ID      int `gorm:"primary_key"`
	User    *User
	UserID  int `gorm:"not null; index" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	Album   *Album
	AlbumID int       `gorm:"not null; index" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	Time    time.Time `sql:"default: null"`
	Count   int
	Length  int
}

type Album struct {
	ID            int `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ModifiedAt    time.Time
	LeftPath      string `gorm:"unique_index:idx_album_abs_path"`
	RightPath     string `gorm:"not null; unique_index:idx_album_abs_path" sql:"default: null"`
	RightPathUDec string `sql:"default: null"`
	Parent        *Album
	ParentID      int       `sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	RootDir       string    `gorm:"unique_index:idx_album_abs_path" sql:"default: null"`
	Genres        []*Genre  `gorm:"many2many:album_genres"`
	Cover         string    `sql:"default: null"`
	Artists       []*Artist `gorm:"many2many:album_artists"`
	TagTitle      string    `sql:"default: null"`
	TagTitleUDec  string    `sql:"default: null"`
	TagBrainzID   string    `sql:"default: null"`
	TagYear       int       `sql:"default: null"`
	Tracks        []*Track
	ChildCount    int `sql:"-"`
	Duration      int `sql:"-"`
	AlbumStar     *AlbumStar
	AlbumRating   *AlbumRating
	AverageRating float64 `sql:"default: null"`
}

func (a *Album) SID() *specid.ID {
	return &specid.ID{Type: specid.Album, Value: a.ID}
}

func (a *Album) ParentSID() *specid.ID {
	return &specid.ID{Type: specid.Album, Value: a.ParentID}
}

func (a *Album) IndexRightPath() string {
	if len(a.RightPathUDec) > 0 {
		return a.RightPathUDec
	}
	return a.RightPath
}

func (a *Album) GenreStrings() []string {
	strs := make([]string, 0, len(a.Genres))
	for _, genre := range a.Genres {
		strs = append(strs, genre.Name)
	}
	return strs
}

func (a *Album) ArtistsStrings() []string {
	var artists = append([]*Artist(nil), a.Artists...)
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].ID < artists[j].ID
	})
	strs := make([]string, 0, len(artists))
	for _, artist := range artists {
		strs = append(strs, artist.Name)
	}
	return strs
}

type PlayQueue struct {
	ID        int `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      *User
	UserID    int `sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	Current   string
	Position  int
	ChangedBy string
	Items     string
}

func (p *PlayQueue) CurrentSID() *specid.ID {
	id, _ := specid.New(p.Current)
	return &id
}

func (p *PlayQueue) GetItems() []specid.ID {
	return splitIDs(p.Items, ",")
}

func (p *PlayQueue) SetItems(items []specid.ID) {
	p.Items = joinIds(items, ",")
}

type TranscodePreference struct {
	User    *User
	UserID  int    `gorm:"not null; unique_index:idx_user_id_client" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	Client  string `gorm:"not null; unique_index:idx_user_id_client" sql:"default: null"`
	Profile string `gorm:"not null" sql:"default: null"`
}

type AlbumArtist struct {
	Album    *Album
	AlbumID  int `gorm:"not null; unique_index:idx_album_id_artist_id" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	Artist   *Artist
	ArtistID int `gorm:"not null; unique_index:idx_album_id_artist_id" sql:"default: null; type:int REFERENCES artists(id) ON DELETE CASCADE"`
}

type TrackGenre struct {
	Track   *Track
	TrackID int `gorm:"not null; unique_index:idx_track_id_genre_id" sql:"default: null; type:int REFERENCES tracks(id) ON DELETE CASCADE"`
	Genre   *Genre
	GenreID int `gorm:"not null; unique_index:idx_track_id_genre_id" sql:"default: null; type:int REFERENCES genres(id) ON DELETE CASCADE"`
}

type AlbumGenre struct {
	Album   *Album
	AlbumID int `gorm:"not null; unique_index:idx_album_id_genre_id" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	Genre   *Genre
	GenreID int `gorm:"not null; unique_index:idx_album_id_genre_id" sql:"default: null; type:int REFERENCES genres(id) ON DELETE CASCADE"`
}

type AlbumStar struct {
	UserID   int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	AlbumID  int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	StarDate time.Time
}

type AlbumRating struct {
	UserID  int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	AlbumID int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES albums(id) ON DELETE CASCADE"`
	Rating  int `gorm:"not null; check:(rating >= 1 AND rating <= 5)"`
}

type ArtistStar struct {
	UserID   int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	ArtistID int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES artists(id) ON DELETE CASCADE"`
	StarDate time.Time
}

type ArtistRating struct {
	UserID   int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	ArtistID int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES artists(id) ON DELETE CASCADE"`
	Rating   int `gorm:"not null; check:(rating >= 1 AND rating <= 5)"`
}

type TrackStar struct {
	UserID   int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	TrackID  int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES tracks(id) ON DELETE CASCADE"`
	StarDate time.Time
}

type TrackRating struct {
	UserID  int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	TrackID int `gorm:"primary_key; not null" sql:"default: null; type:int REFERENCES tracks(id) ON DELETE CASCADE"`
	Rating  int `gorm:"not null; check:(rating >= 1 AND rating <= 5)"`
}

type PodcastAutoDownload string

const (
	PodcastAutoDownloadLatest PodcastAutoDownload = "latest"
	PodcastAutoDownloadNone   PodcastAutoDownload = "none"
)

type Podcast struct {
	ID           int `gorm:"primary_key"`
	UpdatedAt    time.Time
	ModifiedAt   time.Time
	URL          string
	Title        string
	Description  string
	ImageURL     string
	ImagePath    string
	Error        string
	Episodes     []*PodcastEpisode
	AutoDownload PodcastAutoDownload
}

func (p *Podcast) SID() *specid.ID {
	return &specid.ID{Type: specid.Podcast, Value: p.ID}
}

type PodcastEpisodeStatus string

const (
	PodcastEpisodeStatusDownloading PodcastEpisodeStatus = "downloading"
	PodcastEpisodeStatusSkipped     PodcastEpisodeStatus = "skipped"
	PodcastEpisodeStatusDeleted     PodcastEpisodeStatus = "deleted"
	PodcastEpisodeStatusCompleted   PodcastEpisodeStatus = "completed"
	PodcastEpisodeStatusError       PodcastEpisodeStatus = "error"
)

type PodcastEpisode struct {
	ID          int `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ModifiedAt  time.Time
	PodcastID   int `gorm:"not null" sql:"default: null; type:int REFERENCES podcasts(id) ON DELETE CASCADE"`
	Title       string
	Description string
	PublishDate *time.Time
	AudioURL    string
	Bitrate     int
	Length      int
	Size        int
	Path        string
	Filename    string
	Status      PodcastEpisodeStatus
	Error       string
	AbsP        string `gorm:"-"` // TODO: not this. instead we need some consistent way to get the AbsPath for both tracks and podcast episodes. or just files in general
}

func (pe *PodcastEpisode) AudioLength() int  { return pe.Length }
func (pe *PodcastEpisode) AudioBitrate() int { return pe.Bitrate }

func (pe *PodcastEpisode) SID() *specid.ID {
	return &specid.ID{Type: specid.PodcastEpisode, Value: pe.ID}
}

func (pe *PodcastEpisode) PodcastSID() *specid.ID {
	return &specid.ID{Type: specid.Podcast, Value: pe.PodcastID}
}

func (pe *PodcastEpisode) AudioFilename() string {
	return pe.Filename
}

func (pe *PodcastEpisode) Ext() string {
	return filepath.Ext(pe.Filename)
}

func (pe *PodcastEpisode) MIME() string {
	return mime.TypeByExtension(filepath.Ext(pe.Filename))
}

func (pe *PodcastEpisode) AbsPath() string {
	return pe.AbsP
}

type Bookmark struct {
	ID          int `gorm:"primary_key"`
	User        *User
	UserID      int `sql:"default: null; type:int REFERENCES users(id) ON DELETE CASCADE"`
	Position    int
	Comment     string
	EntryIDType string
	EntryID     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type InternetRadioStation struct {
	ID          int `gorm:"primary_key"`
	StreamURL   string
	Name        string
	HomepageURL string
}

func (ir *InternetRadioStation) SID() *specid.ID {
	return &specid.ID{Type: specid.InternetRadioStation, Value: ir.ID}
}
