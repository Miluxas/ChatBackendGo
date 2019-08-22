package models

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func init() {

}

//ChatType is type of chat
type ChatType int

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
var users = []user{user{
	Id:        "admin@e.c",
	firstName: "admin",
	lastName:  "admini",
	username:  "admin",
	password:  "admin",
}, user{
	Id:        "normal@e.c",
	firstName: "normalUser",
	lastName:  "lastNormal",
	username:  "normal",
	password:  "normal",
}, user{
	Id:        "kalim@e.c",
	firstName: "karim",
	lastName:  "Aq Mangool",
	username:  "kalim",
	password:  "kalim",
}, user{
	Id:        "solivan@e.c",
	firstName: "solivan",
	lastName:  "sol",
	username:  "solivan",
	password:  "solivan",
}, user{
	Id:        "ferzin@e.c",
	firstName: "ferzin",
	lastName:  "feriiii",
	username:  "ferzin",
	password:  "ferzin",
}}

type chat struct {
	id          string
	Title       string
	CreateAt    time.Time
	chatType    string
	memberList  []member
	messageList []message
}

func (ch *chat) addMember(newMem *member) {
	var mem member
	isMemberExist := false
	for _, v := range ch.memberList {
		if v.userID == newMem.userID {
			isMemberExist = true
			mem = v
		}
	}

	if !isMemberExist {
		if mem.memberStatus != MEMBER_STATUS__LEFT {
			ch.memberList = append(ch.memberList, *newMem)
		}
	}
}

type message struct {
	id       string
	content  string
	createAt time.Time
	ownerID  string
}

type member struct {
	id           string
	userID       string
	addedAt      time.Time
	memberType   MemberType
	memberStatus MemberStatus
}

type user struct {
	Id        string
	firstName string
	lastName  string
	username  string
	password  string
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
		if v.id == chatID {
			return &ChatList[ind], nil
		}
	}
	return nil, fmt.Errorf("Chat didnt find")
}

//AuthenticateUser authenticate user
func AuthenticateUser(username, password string) string {
	for _, usr := range users {
		if usr.username == username && usr.password == password {
			//currentUser = usr
			return usr.Id
		}
	}
	//currentUser = user{}
	return "__"
}

//StartNewPeerChat start new peer to peer chat with peerUser
func StartNewPeerChat(newChatTitle, currentUserID, userID string) string {
	newMember := member{
		id:           createUniqID(),
		userID:       userID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	ownerMember := member{
		id:           createUniqID(),
		userID:       currentUserID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__OWNER,
		memberStatus: MEMBER_STATUS__NORMAL,
	}
	newChatID := createUniqID()
	newChat := chat{
		id:       newChatID,
		Title:    newChatTitle,
		chatType: Chat_Type__Peer,
		CreateAt: time.Now(),
	}

	newChat.addMember(&ownerMember)
	newChat.addMember(&newMember)

	ChatList = append(ChatList, newChat)
	return newChatID
}

//StartNewGroupChat start new group chat
func StartNewGroupChat(newChatTitle, currentUserID, chatType string) string {

	ownerMember := member{
		id:           createUniqID(),
		userID:       currentUserID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__OWNER,
		memberStatus: MEMBER_STATUS__NORMAL,
	}
	newChatID := createUniqID()
	newChat := chat{
		id:       newChatID,
		Title:    newChatTitle,
		chatType: chatType,
		CreateAt: time.Now(),
	}

	newChat.addMember(&ownerMember)

	ChatList = append(ChatList, newChat)
	return newChatID
}

//SendMessageToChat add message to a chat
func SendMessageToChat(chatID, currentUserID, newMessage string) (string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", err
	}

	newID := createUniqID()
	newMes := message{
		id:       newID,
		content:  newMessage,
		createAt: time.Now(),
		ownerID:  currentUserID,
	}
	chat.messageList = append(chat.messageList, newMes)

	return newID, nil
}

//JoinToChat join current user to a chat
func JoinToChat(chatID, currentUserID string) (string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", err
	}
	newID := createUniqID()
	newMember := member{
		id:           newID,
		userID:       currentUserID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	if chat.chatType == Chat_Type__PRIVATE_CANNAL || chat.chatType == Chat_Type__PRIVATE_GROUP {
		newMember.memberStatus = MEMBER_STATUS__REQUESTED
	}

	chat.addMember(&newMember)
	return newID, nil
}

//AddOtherUserToChat add other user to a chat
func AddOtherUserToChat(chatID, currentUserID, userID string) (string, error) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return "", err
	}
	newID := createUniqID()
	newMember := member{
		id:           newID,
		userID:       userID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	chat.addMember(&newMember)
	return newID, nil
}

//GetChat return a chat as json byte array
func GetChat(chatID, currentUserID string) ([]byte, error) {
	var eb []byte
	chat, err := getChatFromID(chatID)
	if err != nil {
		return eb, err
	}
	isUserMemberOfChat := false
	for _, v := range chat.memberList {
		if v.userID == currentUserID {
			isUserMemberOfChat = true
			break
		}
	}
	if !isUserMemberOfChat {
		return eb, fmt.Errorf("User isn't member of chat")
	}
	//fmt.Println(chat, *chat)
	jChat, err := json.Marshal(*chat)
	fmt.Println(string(jChat))
	if err != nil {
		return eb, err
	}
	return jChat, nil
}
