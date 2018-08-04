package main

import (
	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var conf Config

func main() {
	tomlData, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		panic("can't read config.toml")
	}
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		panic("config parse error")
	}
	log.Println("line")
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "DELETE"}
	router.Use(cors.New(config))

	//serve for music fils.
	router.Static("/MusicServer/file/", conf.File.Root)

	//serve for files list.
	router.GET("/MusicServer/dir", directoryHandler)

	/////MONGOBD
	router.GET("/MusicServer/songlist", showSongListHandler)
	router.GET("/MusicServer/songlist/:listname", singleSongListHandler)
	router.POST("/MusicServer/songlist", addToSongListHandler)
	router.POST("/MusicServer/songquery", songQueryHandler)
	router.DELETE("/MusicServer/songlist", deleteSongHandler)
	/////

	router.Run(":8026")
	log.Println("Serveing on 8026")
}

func deleteSongHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	songList := c.PostForm("songlist")
	songname := c.PostForm("name")
	songurl := c.PostForm("url")
	// Collection
	collection := session.DB(conf.DB.Name).C(songList)

	deleteSong := Song{
		Name: songname,
		Url:  songurl,
	}

	// delete
	if _, err := collection.RemoveAll(deleteSong); err != nil {
		panic(err)
	}

	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

}
func songQueryHandler(c *gin.Context) {
	songurl := c.PostForm("url")
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	songListOutput := []string{}
	var songs []Song
	for _, songListName := range songListNames {
		err = session.DB(conf.DB.Name).C(songListName).Find(bson.M{"url": songurl}).All(&songs)
		if err != nil {
			panic(err)
		}

		if len(songs) != 0 {
			songListOutput = append(songListOutput, songListName)
			log.Println("findin: " + songListName)
		}

	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListOutput,
	})

	c.JSON(http.StatusOK, list)

}

func showSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

}
func singleSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Collection
	listName := c.Param("listname")

	collection := session.DB(conf.DB.Name).C(listName)

	// Find All
	var songs []Song
	err = collection.Find(nil).All(&songs)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, songs)

}
func addToSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	songList := c.PostForm("songlist")
	songname := c.PostForm("name")
	songurl := c.PostForm("url")
	// Collection
	collection := session.DB(conf.DB.Name).C(songList)

	insertSong := Song{
		Name: songname,
		Url:  songurl,
	}
	log.Println(insertSong)
	// Insert
	if err := collection.Insert(insertSong); err != nil {
		panic(err)
	}

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

}

func directoryHandler(c *gin.Context) {
	dir := c.Query("dir")
	//read files in directory.
	files, err := ioutil.ReadDir(conf.File.Root + dir)
	if err != nil {

	}
	s_dir := make([]os.FileInfo, 0)  //list of folders
	s_file := make([]os.FileInfo, 0) //list of files

	for _, f := range files {
		if f.Name()[0] != []byte(".")[0] {
			ext := filepath.Ext(f.Name())
			if f.IsDir() {
				s_dir = append(s_dir, f)
			} else if isAudioExt(ext) { //check if file is music file.
				s_file = append(s_file, f)
			}
		}
	}
	s_dir = append(s_dir, s_file...)

	//Make the list of json for output.
	names := make([]Item, 0)
	for _, f := range s_dir {
		names = append(names, Item{
			Name:  f.Name(),
			IsDir: f.IsDir(),
		})
	}
	c.JSON(http.StatusOK, names)
}

func isAudioExt(val string) bool {
	for i := range conf.File.AudioExt {
		if conf.File.AudioExt[i] == val {
			return true
		}
	}
	return false
}

// global config type
type Config struct {
	File fileConfig `toml:"file"`
	DB   dbConfig   `toml:"database"`
}

type fileConfig struct {
	Root     string   `toml:"root"`
	AudioExt []string `toml:"audioExt"`
}

type dbConfig struct {
	Host []string `toml:"host"`
	Name string   `toml:"dbName"`
}

//datatype of file or folder
type Item struct {
	Action string //joystick
	Name   string
	IsDir  bool
}

//datatype of song in db.
type Song struct {
	Name string
	Url  string
}
type SongInDB struct {
	ID   bson.ObjectId
	Name string
	Url  string
}
type SongUrl struct {
	Url string
}

type SongListAll struct {
	SongListNames []string
}
