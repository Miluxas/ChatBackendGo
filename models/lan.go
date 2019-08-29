package models

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustin/go-broadcast"
)

func init() {

}

//MemberType is type of member in chat
type MemberType int

//MemberStatus is status of member in chat
type MemberStatus int

const (
	Chat_Type__Peer           string = "PEER"
	Chat_Type__PUBLIC_GROUP   string = "PUBLIC_GROUP"
	Chat_Type__PRIVATE_GROUP  string = "PRIVATE_GROUP"
	Chat_Type__PUBLIC_CANNAL  string = "PUBLIC_CANNAL"
	Chat_Type__PRIVATE_CANNAL string = "PRIVATE_CANNAL"

	MEMBER_TYPE__OWNER  MemberType = 0
	MEMBER_TYPE__ADMIN  MemberType = 1
	MEMBER_TYPE__NORMAL MemberType = 2

	MEMBER_STATUS__NORMAL    MemberStatus = 0
	MEMBER_STATUS__BLOCKED   MemberStatus = 1
	MEMBER_STATUS__REQUESTED MemberStatus = 2
	MEMBER_STATUS__LEFT      MemberStatus = 3
	MEMBER_STATUS__EXPELED   MemberStatus = 4
)

var ChatList []chat
var users = []User{User{
	ID:        "admin@e.c",
	FirstName: "admin",
	LastName:  "admini",
	username:  "admin",
	password:  "admin",
}, User{
	ID:        "normal@e.c",
	FirstName: "normalUser",
	LastName:  "lastNormal",
	username:  "normal",
	password:  "normal",
}, User{
	ID:        "kalim@e.c",
	FirstName: "karim",
	LastName:  "Aq Mangool",
	username:  "kalim",
	password:  "kalim",
}, User{
	ID:        "solivan@e.c",
	FirstName: "solivan",
	LastName:  "sol",
	username:  "solivan",
	password:  "solivan",
}, User{
	ID:        "ferzin@e.c",
	FirstName: "ferzin",
	LastName:  "feriiii",
	username:  "ferzin",
	password:  "ferzin",
}}

type chat struct {
	ID          string
	Title       string
	CreateAt    time.Time
	ChatType    string
	MemberList  []member
	MessageList []message
}

func (ch *chat) addMember(newMem *member) {
	var mem member
	isMemberExist := false
	for _, v := range ch.MemberList {
		if v.UserID == newMem.UserID {
			isMemberExist = true
			mem = v
		}
	}

	if !isMemberExist {
		if mem.MemberStatus != MEMBER_STATUS__LEFT {
			ch.MemberList = append(ch.MemberList, *newMem)
		}
	}
}

func (ch *chat) findMember(userID string) bool {
	for _, v := range ch.MemberList {
		if v.UserID == userID && v.MemberStatus == MEMBER_STATUS__NORMAL {
			//fmt.Println(v)
			return true
		}
	}
	return false
}

type message struct {
	ID       string
	Content  string
	CreateAt time.Time
	OwnerID  string
}

type member struct {
	ID           string
	UserID       string
	AddedAt      time.Time
	MemberType   MemberType
	MemberStatus MemberStatus
}

//User user of the chat system
type User struct {
	ID        string
	FirstName string
	LastName  string
	username  string
	password  string
}

//Alert alert for realtime
type Alert struct {
	AlertType string
	Data      interface{}
}

//var currentUser user

func createUniqID() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return ""
	}

	u[8] = (u[8] | 0x80) & 0xBF // what does this do?
	u[6] = (u[6] | 0x40) & 0x4F // what does this do?

	return hex.EncodeToString(u)
}

func getChatFromID(chatID string) (*chat, error) {
	for ind, v := range ChatList {
		if v.ID == chatID {
			return &ChatList[ind], nil
		}
	}
	return nil, fmt.Errorf("Chat didnt find")
}

//AuthenticateUser authenticate user
func AuthenticateUser(username, password string) User {
	for _, usr := range users {
		if usr.username == username && usr.password == password {
			//currentUser = usr
			return usr
		}
	}
	//currentUser = User{}
	return User{}
}

//StartNewPeerChat start new peer to peer chat with peerUser
func StartNewPeerChat(newChatTitle, currentUserID, userID string) (string, error) {

	for _, v := range ChatList {
		if v.ChatType == Chat_Type__Peer && ((v.MemberList[0].UserID == userID && v.MemberList[1].UserID == currentUserID) ||
			(v.MemberList[1].UserID == userID && v.MemberList[0].UserID == currentUserID)) {
			return "", fmt.Errorf("peer chat with this member is exist")
		}
	}

	newMember := member{
		ID:           createUniqID(),
		UserID:       userID,
		AddedAt:      time.Now(),
		MemberType:   MEMBER_TYPE__NORMAL,
		MemberStatus: MEMBER_STATUS__NORMAL,
	}

	ownerMember := member{
		ID:           createUniqID(),
		UserID:       currentUserID,
		AddedAt:      time.Now(),
		MemberType:   MEMBER_TYPE__OWNER,
		MemberStatus: MEMBER_STATUS__NORMAL,
	}
	newChatID := createUniqID()
	newChat := chat{
		ID:       newChatID,
		Title:    newChatTitle,
		ChatType: Chat_Type__Peer,
		CreateAt: time.Now(),
	}

	newChat.addMember(&ownerMember)
	newChat.addMember(&newMember)

	ChatList = append(ChatList, newChat)
	return newChatID, nil
}

//StartNewGroupChat start new group chat
func StartNewGroupChat(newChatTitle, currentUserID, chatType string) string {

	ownerMember := member{
		ID:           createUniqID(),
		UserID:       currentUserID,
		AddedAt:      time.Now(),
		MemberType:   MEMBER_TYPE__OWNER,
		MemberStatus: MEMBER_STATUS__NORMAL,
	}
	newChatID := createUniqID()
	newChat := chat{
		ID:       newChatID,
		Title:    newChatTitle,
		ChatType: chatType,
		CreateAt: time.Now(),
	}

	newChat.addMember(&ownerMember)

	ChatList = append(ChatList, newChat)
	return newChatID
}

//SendMessageToChat add message to a chat
func SendMessageToChat(chatID, currentUserID, newMessage string) (time.Time, string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return time.Now(), "", err
	}

	newID := createUniqID()
	cAt := time.Now()
	newMes := message{
		ID:       newID,
		Content:  newMessage,
		CreateAt: cAt,
		OwnerID:  currentUserID,
	}
	chat.MessageList = append(chat.MessageList, newMes)

	return cAt, newID, nil
}

//JoinToChat join current user to a chat
func JoinToChat(chatID, currentUserID string) (string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", err
	}
	newID := createUniqID()
	newMember := member{
		ID:           newID,
		UserID:       currentUserID,
		AddedAt:      time.Now(),
		MemberType:   MEMBER_TYPE__NORMAL,
		MemberStatus: MEMBER_STATUS__NORMAL,
	}

	if chat.ChatType == Chat_Type__PRIVATE_CANNAL || chat.ChatType == Chat_Type__PRIVATE_GROUP {
		newMember.MemberStatus = MEMBER_STATUS__REQUESTED
	}

	chat.addMember(&newMember)
	return newID, nil
}

//LeaveChat leave user from a chat
func LeaveChat(chatID, currentUserID string) (string, string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", "", err
	}
	for ind, v := range chat.MemberList {
		if v.UserID == currentUserID {
			chat.MemberList[ind].MemberStatus = MEMBER_STATUS__LEFT
			//fmt.Println(chat)
			return v.UserID, v.ID, nil
		}
	}
	fmt.Println(chat)

	return "", "", nil
}

//BlockPeerChat leave user from a chat
func BlockPeerChat(chatID, currentUserID string) (string, string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", "", err
	}
	if chat.ChatType != Chat_Type__Peer {
		return "", "", fmt.Errorf("not a peer chat")
	}
	for ind, v := range chat.MemberList {
		if v.UserID != currentUserID {
			chat.MemberList[ind].MemberStatus = MEMBER_STATUS__BLOCKED
			//fmt.Println(chat)
			return v.UserID, v.ID, nil
		}
	}

	return "", "", nil
}

//AddOtherUserToChat add other user to a chat
func AddOtherUserToChat(chatID, currentUserID, userID string) (string, string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", "", err
	}
	newID := createUniqID()
	newMember := member{
		ID:           newID,
		UserID:       userID,
		AddedAt:      time.Now(),
		MemberType:   MEMBER_TYPE__NORMAL,
		MemberStatus: MEMBER_STATUS__NORMAL,
	}

	chat.addMember(&newMember)
	return chat.Title, newID, nil
}

//SendAlertToMember send a alert to all member of chat
func SendAlertToMember(chatID string, newAlert interface{}) {
	chat, err := getChatFromID(chatID)
	if err != nil {
		return
	}
	for _, v := range chat.MemberList {
		if v.MemberStatus == MEMBER_STATUS__NORMAL {
			UserChannel(v.UserID).Submit(newAlert)
		}
	}

}

//SendAlertToOneMember send a alert to a member
func SendAlertToOneMember(userID string, newAlert interface{}) {
	UserChannel(userID).Submit(newAlert)
}

//GetChat return a chat as json byte array
func GetChat(chatID, currentUserID string) (string, error) {
	chatf, err := getChatFromID(chatID)
	if err != nil {
		return "", err
	}
	chat := chatf

	isUserMemberOfChat := false
	for _, v := range chat.MemberList {
		if v.UserID == currentUserID {
			isUserMemberOfChat = true
			break
		}
	}
	if !isUserMemberOfChat {
		return "", fmt.Errorf("User isn't member of chat")
	}
	if chat.ChatType == Chat_Type__Peer {
		if chat.MemberList[0].UserID == currentUserID {
			chat.Title = chat.MemberList[1].UserID
		} else {
			chat.Title = chat.MemberList[0].UserID
		}
	}
	//fmt.Println(chat, *chat)
	jChat, err := json.Marshal(*chat)
	//fmt.Println(string(jChat))
	if err != nil {
		return "", err
	}
	return string(jChat), nil
}

//GetChatList return user chat list as json byte array
func GetChatList(currentUserID string) (string, error) {
	var tmpList []chat
	for _, v := range ChatList {
		if v.findMember(currentUserID) {

			tmpList = append(tmpList, v)
		}
	}
	//fmt.Println(chat, *chat)
	jChat, err := json.Marshal(tmpList)
	//fmt.Println(string(jChat))
	if err != nil {
		return "", err
	}
	return string(jChat), nil
}

/********************************************************************/
/*					realtime functions								*/
/*																	*/
/********************************************************************/
var userChannels = make(map[string]broadcast.Broadcaster)

//OpenListener open listener
func OpenListener(userid string) chan interface{} {
	listener := make(chan interface{})
	UserChannel(userid).Register(listener)
	return listener
}

//CloseListener close listener
func CloseListener(userid string, listener chan interface{}) {
	UserChannel(userid).Unregister(listener)
	close(listener)
}

//DeleteBroadcast delete broadcast
func DeleteBroadcast(userid string) {
	b, ok := userChannels[userid]
	if ok {
		b.Close()
		delete(userChannels, userid)
	}
}

//UserChannel get user channel
func UserChannel(userid string) broadcast.Broadcaster {
	b, ok := userChannels[userid]
	if !ok {
		b = broadcast.NewBroadcaster(10)
		userChannels[userid] = b
	}
	return b
}
