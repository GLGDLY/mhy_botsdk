package api_models

type APIBaseModel struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	RawData string // 原始json数据
}

type EmptyModel struct {
	APIBaseModel
	Data struct{} `json:"data"`
}

/* MemberAccess */

type CheckMemberBotAccessTokenModel struct {
	APIBaseModel
	Data struct {
		AccessInfo struct {
			UID               string `json:"uid"`
			VillaID           string `json:"villa_id"`
			MemberAccessToken string `json:"member_access_token"`
			BotTplID          string `json:"bot_tpl_id"`
		} `json:"access_info"`
		Member struct {
			Basic struct {
				UID       string `json:"uid"`
				Nickname  string `json:"nickname"`
				Introduce string `json:"introduce"`
				Avatar    string `json:"avatar"`
				AvatarURL string `json:"avatar_url"`
			} `json:"basic"`
			RoleIDList []string `json:"role_id_list"`
			JoinedAt   string   `json:"joined_at"`
			RoleList   []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Color    string `json:"color"`
				RoleType string `json:"role_type"`
				VillaID  string `json:"villa_id"`
			} `json:"role_list"`
		} `json:"member"`
	} `json:"data"`
}

/* Villa */

type GetVillaModel struct {
	APIBaseModel
	Data struct {
		Villa struct {
			VillaID        uint64   `json:"villa_id,string"`
			Name           string   `json:"name"`
			VillaAvatarURL string   `json:"villa_avatar_url"`
			OwnerUID       uint64   `json:"owner_uid,string"`
			IsOfficial     bool     `json:"is_official"`
			Introduce      string   `json:"introduce"`
			CategoryID     uint32   `json:"category_id"`
			Tags           []string `json:"tags"`
		} `json:"villa"`
	} `json:"data"`
}

/* Member */

type MemberRoleModel struct {
	ID        uint64            `json:"id,string"`
	Name      string            `json:"name"`
	VillaID   uint64            `json:"villa_id,string"`
	Color     string            `json:"color"`
	RoleType  string            `json:"role_type"`
	IsAllRoom bool              `json:"is_all_room"`
	RoomIDs   Uint64StringSlice `json:"room_ids"`
}

type MemberModel struct {
	Basic struct {
		UID       string `json:"uid"`
		Nickname  string `json:"nickname"`
		Introduce string `json:"introduce"`
		AvatarURL string `json:"avatar_url"`
	} `json:"basic"`
	RoleIDList []string          `json:"role_id_list"`
	JoinedAt   string            `json:"joined_at"`
	RoleList   []MemberRoleModel `json:"role_list"`
}

type GetMemberModel struct {
	APIBaseModel
	Data struct {
		Member MemberModel `json:"member"`
	} `json:"data"`
}

type GetVillaMembersModel struct {
	APIBaseModel
	Data struct {
		List          []MemberModel `json:"list"`
		NextOffsetStr string        `json:"next_offset_str"`
	} `json:"data"`
}

/* Message */

type SendMessageModel struct {
	APIBaseModel
	Data struct {
		BotMsgId string `json:"bot_msg_id"`
	} `json:"data"`
}

/* Room */

type CreateRoomModel struct {
	APIBaseModel
	Data struct {
		GoomID uint64 `json:"group_id,string"`
	} `json:"data"`
}

type GetGroupListModel struct {
	APIBaseModel
	Data struct {
		List []struct {
			GroupID   uint64 `json:"group_id,string"`
			GroupName string `json:"group_name"`
		} `json:"list"`
	} `json:"data"`
}

type GetRoomModel struct {
	APIBaseModel
	Data struct {
		Room struct {
			RoomID                uint64 `json:"room_id,string"`
			RoomName              string `json:"room_name"`
			RoomType              string `json:"room_type"`
			GroupID               uint64 `json:"group_id,string"`
			RoomDefaultNotifyType string `json:"room_default_notify_type"`
			SendMsgAuthRange      struct {
				IsAllSendMsg bool              `json:"is_all_send_msg"`
				Roles        Uint64StringSlice `json:"roles"`
			} `json:"send_msg_auth_range"`
		} `json:"room"`
	} `json:"data"`
}

type GetVillaGroupRoomListModel struct {
	APIBaseModel
	Data struct {
		GroupID   uint64 `json:"group_id,string"`
		GroupName string `json:"group_name"`
		RoomList  []struct {
			RoomID   uint64 `json:"room_id,string"`
			RoomName string `json:"room_name"`
			RoomType string `json:"room_type"`
			GroupID  uint64 `json:"group_id,string"`
		} `json:"room_list"`
	} `json:"data"`
}

/* Role */

type CreateRoleModel struct {
	APIBaseModel
	Data struct {
		ID uint64 `json:"id,string"`
	} `json:"data"`
}

type GetRoleInfoModel struct {
	APIBaseModel
	Data struct {
		Role struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Color       string `json:"color"`
			VillaID     string `json:"villa_id"`
			RoleType    string `json:"role_type"`
			MemberNum   string `json:"member_num"`
			Permissions []struct {
				Key      string `json:"key"`
				Name     string `json:"name"`
				Describe string `json:"describe"`
			} `json:"permissions"`
		} `json:"role"`
	} `json:"data"`
}

type GetVillaMemberRolesModel struct {
	APIBaseModel
	Data struct {
		List []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Color     string `json:"color"`
			RoleType  string `json:"role_type"`
			VillaID   string `json:"villa_id"`
			MemberNum string `json:"member_num"`
		} `json:"list"`
	} `json:"data"`
}

/* Emoticon */

type GetAllEmoticonsModel struct {
	APIBaseModel
	Data struct {
		List []struct {
			EmoticonID   uint64 `json:"emoticon_id,string"`
			DescribeText string `json:"describe_text"`
			Icon         string `json:"icon"`
		} `json:"list"`
	} `json:"data"`
}

/* Audit */

type AuditModel struct {
	APIBaseModel
	Data struct {
		AuditID string `json:"audit_id"`
	} `json:"data"`
}

/* Image */

type UploadImageModel struct {
	APIBaseModel
	Data struct {
		NewURL string `json:"new_url"`
	} `json:"data"`
}
