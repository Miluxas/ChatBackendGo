package models

import (
	"crypto/rand"
	"encoding/hex"
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
	Chat_Type__Peer           ChatType = 0
	Chat_Type__PUBLIC_GROUP   ChatType = 1
	Chat_Type__PRIVATE_GROUP  ChatType = 2
	Chat_Type__PUBLIC_CANNAL  ChatType = 3
	Chat_Type__PRIVATE_CANNAL ChatType = 4

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
}, user{
	Id:        "normal@e.c",
	firstName: "normalUser",
	lastName:  "lastNormal",
}, user{
	Id:        "kalim@e.c",
	firstName: "karim",
	lastName:  "Aq Mangool",
}, user{
	Id:        "solivan@e.c",
	firstName: "solivan",
	lastName:  "sol",
}, user{
	Id:        "ferzin@e.c",
	firstName: "ferzin",
	lastName:  "feriiii",
}}

type chat struct {
	id          string
	title       string
	createAt    time.Time
	chatType    ChatType
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
}

var currentUser user

func SetCurrentUser(userID string) {
	for _, usr := range users {
		if usr.Id == userID {
			currentUser = usr
		}
	}
}

func getChatFromID(chatID string) (*chat, error) {
	for ind, v := range ChatList {
		if v.id == chatID {
			return &ChatList[ind], nil
		}
	}
	return nil, fmt.Errorf("Chat didnt find")
}

//StartNewPeerChat start new peer to peer chat with peerUser
func StartNewPeerChat(newChatID string, newChatTitle string, userID string) {
	newMember := member{
		id:           createUniqID(),
		userID:       userID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	ownerMember := member{
		id:           createUniqID(),
		userID:       currentUser.Id,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__OWNER,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	newChat := chat{
		id:       newChatID,
		title:    newChatTitle,
		chatType: Chat_Type__Peer,
		createAt: time.Now(),
	}

	newChat.addMember(&ownerMember)
	newChat.addMember(&newMember)

	ChatList = append(ChatList, newChat)
}

//StartNewGroupChat start new group chat
func StartNewGroupChat(newChatID string, newChatTitle string, chatType ChatType) {

	ownerMember := member{
		id:           createUniqID(),
		userID:       currentUser.Id,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__OWNER,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	newChat := chat{
		id:       newChatID,
		title:    newChatTitle,
		chatType: chatType,
		createAt: time.Now(),
	}

	newChat.addMember(&ownerMember)

	ChatList = append(ChatList, newChat)
}

//SendMessageToChat add message to a chat
func SendMessageToChat(chatID string, newMessage string) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return
	}

	newMes := message{
		id:       createUniqID(),
		content:  newMessage,
		createAt: time.Now(),
		ownerID:  currentUser.Id,
	}
	chat.messageList = append(chat.messageList, newMes)
}

//JoinToChat join current user to a chat
func JoinToChat(chatID string) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return
	}

	newMember := member{
		id:           createUniqID(),
		userID:       currentUser.Id,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	if chat.chatType == Chat_Type__PRIVATE_CANNAL || chat.chatType == Chat_Type__PRIVATE_GROUP {
		newMember.memberStatus = MEMBER_STATUS__REQUESTED
	}

	chat.addMember(&newMember)

}

//AddOtherUserToChat add other user to a chat
func AddOtherUserToChat(chatID string, userID string) {

	chat, err := getChatFromID(chatID)
	if err != nil {
		return
	}

	newMember := member{
		id:           createUniqID(),
		userID:       userID,
		addedAt:      time.Now(),
		memberType:   MEMBER_TYPE__NORMAL,
		memberStatus: MEMBER_STATUS__NORMAL,
	}

	chat.addMember(&newMember)

}

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
