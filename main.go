package main

import(
	//"html/template"
	"log"
	"io/ioutil"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	//"github.com/gorilla/websocket"
	// "encoding/json"
	//"strconv"
	"os"
	"path/filepath"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)
var audioExt = map[string]bool{
	".wav": true,".webm": true,".opus": true,".ogg": true,".mp3": true,".m4a": true,".flac": true,
}
var root = "./web/mnt"
/////const for MONGODB.
var Host = []string{
	"127.0.0.1:27017",
	// replica set addrs...
}
const (
	Username   = "YOUR_USERNAME"
	Password   = "YOUR_PASS"
	Database   = "playerSongList"
	Collection = "YOUR_COLLECTION"
)
/////
func main(){
	router := gin.Default()
	//router.GET("/ws/MusicPlayer", func(c *gin.Context){wshandler(c.Writer, c.Request)})
	router.Use(cors.Default())

	//serve for music fils.
	router.Static("/MusicServer/file/", root)

	//serve for files list.
	router.GET("/MusicServer/dir", directoryHandler)

	
	/////MONGOBD
	router.GET("/MusicServer/songlist", showSongListHandler)
	router.GET("/MusicServer/songlist/:listname", singleSongListHandler)
	router.POST("/MusicServer/songlist", addToSongListHandler)
	router.POST("/MusicServer/songquery", songQueryHandler)
	/////


	router.Run(":8026")
	log.Println("Serveing on 8026")
}
func songQueryHandler(c *gin.Context){
//	songurl := c.Query("url")
//	log.Println("songurl="+songurl)
	songurl := c.PostForm("url")
//	log.Println("songurl2="+songurl2)
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: Host,
		// Username: Username,
		// Password: Password,
		// Database: Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()


//	querySong := SongUrl{
//		Url: songurl,
//	}
//	//SongLists in DB.
	songListNames, err := session.DB(Database).CollectionNames()
	if err != nil{
		panic(err)
	}

	songListOutput := []string{}
	var songs []Song
	for _, songListName := range songListNames {
		err = session.DB(Database).C(songListName).Find(bson.M{"url": songurl}).All(&songs)
		if err != nil{
			panic(err)
		}

		if(len(songs)!=0){
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

func showSongListHandler(c *gin.Context){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: Host,
		// Username: Username,
		// Password: Password,
		// Database: Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//SongLists in DB.
	songListNames, err := session.DB(Database).CollectionNames()
	if err != nil{
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)
	
	list = append(list, SongListAll{
		SongListNames: songListNames,
	})
	
	c.JSON(http.StatusOK, list)


}
func singleSongListHandler(c *gin.Context){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: Host,
		// Username: Username,
		// Password: Password,
		// Database: Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Collection
	listName := c.Param("listname")
	
	collection := session.DB(Database).C(listName)

	// Find All
	var songs []Song
	err = collection.Find(nil).All(&songs)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, songs)
	

}
func addToSongListHandler(c *gin.Context){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: Host,
		// Username: Username,
		// Password: Password,
		// Database: Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()


	songList := c.PostForm("songlist")
	songname := c.PostForm("name")
	songurl := c.PostForm("url")
	// Collection
	collection := session.DB(Database).C(songList)

	insertSong := Song{
		Name:songname,
		Url: songurl,
	}
	log.Println(insertSong)
	// Insert
	if err := collection.Insert(insertSong); err != nil {
		panic(err)
	}

//	// Collection
//	collection := session.DB(Database).C("newnewbe")
//
//	insertSong := Song{
//		Name:"haha"  ,
//		Url: "hahaurl",
//	}
//
//	// Insert
//	if err := collection.Insert(insertSong); err != nil {
//		panic(err)
//	}
//
	//SongLists in DB.
	songListNames, err := session.DB(Database).CollectionNames()
	if err != nil{
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)
	
	list = append(list, SongListAll{
		SongListNames: songListNames,
	})
	
	c.JSON(http.StatusOK, list)

}


func directoryHandler(c *gin.Context){
	dir := c.Query("dir")
	//read files in directory.
	files, err := ioutil.ReadDir(root + dir)
	if err != nil {

	}
	s_dir := make([]os.FileInfo,0)    //list of folders
	s_file := make([]os.FileInfo,0)   //list of files
	
	for _, f := range files{
		if(f.Name()[0] != []byte(".")[0]){
			ext := filepath.Ext(f.Name())
			if(f.IsDir()){
				s_dir = append(s_dir, f)
			} else if audioExt[ext]{ //check if file is music file. 
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

	/////WEBSOCKET VERSION
	// for _, f := range s_dir{
	// 	msg_out := Item{
	// 		"item",
	// 		f.Name(),
	// 		strconv.FormatBool(f.IsDir()),
	// 	}
	// 	conn.WriteJSON(msg_out)
	// }
	/////
}


/////////////////WEBSOCKET VERSION//////////////////
// var upgrader =websocket.Upgrader{
// 	CheckOrigin :func(r *http.Request)bool{
// 		return true
// 	},
// }
// func wshandler(w http.ResponseWriter, r *http.Request){
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	//defer conn.Close()
// 	if err != nil{
// 		log.Println("Failed to set websocket upgrade: %+v", err)
// 		return
// 	}
// 	var msg map[string]interface{}
// 	for{
// 		err = conn.ReadJSON(&msg)
// 		if err != nil{
// 			log.Println("Failed to read json", err)
// 			return
// 		}
// 		log.Println(msg["Action"].(string))
// 		switch msg["Action"].(string){
// 		case "list":
// 			log.Println("Got list")
// 			msg_out := UpdateList{
// 				"list",
// 			}
// 			log.Println(msg_out);
// 			conn.WriteJSON(msg_out)
// 			//start read file
// 			files, err := ioutil.ReadDir(root+msg["Path"].(string))
// 			log.Println(root+msg["Path"].(string));
// 			if err != nil{
// 				log.Fatal(err)
// 			}
// 			s_dir := make([]os.FileInfo,0)
// 			s_file := make([]os.FileInfo,0)
// 			for _,f:=range files{
// 				if(f.Name()[0] != []byte(".")[0]){
// 					ext := filepath.Ext(f.Name())
// 					if(f.IsDir()){
// 						s_dir=append(s_dir,f)
// 					} else if audioExt[ext]{
// 						s_file = append(s_file,f)
// 					}
// 				}
// 			}
// 			s_dir = append(s_dir, s_file...)
// 			for _,f:=range s_dir{
// 				msg_out := Item{
// 					"item",
// 					f.Name(),
// 					strconv.FormatBool(f.IsDir()),
// 				}
// 				conn.WriteJSON(msg_out)
// 			}
// 			msg_out = UpdateList{
// 				"end",
// 			}
// 			log.Println(msg_out);
// 			conn.WriteJSON(msg_out)
// 		}
// 	}
//
// }
// type UpdateList struct{
// 	Action string
// }
////////////////////////////////////////////////

//datatype of file or folder
type Item struct{
	Action string//joystick
	Name string
	IsDir bool
}


//datatype of song in db.
type Song struct{
	Name string
	Url string
}
type SongInDB struct{
	ID bson.ObjectId
	Name string
	Url string
}
type SongUrl struct{
	Url string
}

type SongListAll struct{
	SongListNames []string
}
