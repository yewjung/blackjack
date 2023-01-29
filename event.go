package main

type Action int

const (
	JOIN_ROOM Action = iota
	LEAVE_ROOM
	CREATE_ROOM
	START_GAME
)

type Event int

const (
	PLAYER_ADDED Event = iota
	ROOM_CREATED
	JOINED_ROOM
	LEFT_ROOM
	ERROR
)

type Error int

const (
	ROOM_NOT_EXIST Error = iota
	ROOM_ID_NOT_PROVIDED
	ROOM_ID_FORMAT_WRONG
	CANNOT_CREATE_ROOM
	CANNOT_JOIN_ROOM
)
