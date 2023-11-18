package mgmt_pb

func (request *InspectRequest) GetContentType() int32 {
	return int32(ContentType_InspectRequestType)
}

func (request *InspectResponse) GetContentType() int32 {
	return int32(ContentType_InspectResponseType)
}

func (request *SshTunnelRequest) GetContentType() int32 {
	return int32(ContentType_SshTunnelRequestType)
}

func (request *SshTunnelResponse) GetContentType() int32 {
	return int32(ContentType_SshTunnelResponseType)
}

func (request *RaftMemberListResponse) GetContentType() int32 {
	return int32(ContentType_RaftListMembersResponseType)
}
