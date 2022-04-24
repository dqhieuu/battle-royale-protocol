package utils

const (
	PacketType_ClientConnectToLobbyServer      = 1
	PacketType_GameServerConnectToLobbyServer  = 2
	PacketType_LobbyServerToClient             = 3
	PacketType_GameServerPeriodicToLobbyServer = 4
	PacketType_ClientConnectToGameServer       = 5
	PacketType_GameServerAndClientInGame       = 6
)

const (
	PacketSubType_ClientConnectToLobbyServer_LoginRequest          = 1
	PacketSubType_ClientConnectToLobbyServer_LoginSuccess          = 2
	PacketSubType_ClientConnectToLobbyServer_LoginFailure          = 3
	PacketSubType_GameServerConnectToLobbyServer_AddToLobby        = 1
	PacketSubType_GameServerConnectToLobbyServer_AddToLobbySuccess = 2
	PacketSubType_GameServerConnectToLobbyServer_AddToLobbyFailure = 3
	PacketSubType_LobbyServerToClient_GameServerAddress            = 1
	PacketSubType_LobbyServerToClient_ForceDisconnect              = 2
	PacketSubType_GameServerPeriodicToLobbyServer_RoomStatus       = 1
	PacketSubType_ClientConnectToGameServer_JoinRoom               = 1
	PacketSubType_ClientConnectToGameServer_JoinRoomSuccess        = 2
	PacketSubType_ClientConnectToGameServer_JoinRoomFailure        = 3
	PacketSubType_GameServerAndClientInGame_PositionUpdate         = 1
	PacketSubType_GameServerAndClientInGame_ItemUpdate             = 2
	PacketSubType_GameServerAndClientInGame_HPUpdate               = 3
	PacketSubType_GameServerAndClientInGame_StateUpdate            = 4
)

const (
	PacketSubType_ClientConnectToLobbyServer_JoinRoomSucess_WithoutRoomInfo = 0
	PacketSubType_ClientConnectToLobbyServer_JoinRoomSucess_WithRoomInfo    = 1
)
