package helix

import (
	"strconv"
	"strings"
	"time"
)

// parseIRCMessage parses a raw IRC message into an IRCMessage struct.
// Format: [@tags] [:prefix] COMMAND [params...] [:trailing]
func parseIRCMessage(raw string) *IRCMessage {
	msg := &IRCMessage{
		Raw:  raw,
		Tags: make(map[string]string),
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return msg
	}

	pos := 0

	// Parse tags (starts with @)
	if raw[pos] == '@' {
		end := strings.Index(raw, " ")
		if end == -1 {
			return msg
		}
		msg.Tags = parseTags(raw[1:end])
		pos = end + 1
	}

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse prefix (starts with :)
	if pos < len(raw) && raw[pos] == ':' {
		end := strings.Index(raw[pos:], " ")
		if end == -1 {
			msg.Prefix = raw[pos+1:]
			return msg
		}
		msg.Prefix = raw[pos+1 : pos+end]
		pos = pos + end + 1
	}

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse command
	end := strings.Index(raw[pos:], " ")
	if end == -1 {
		msg.Command = raw[pos:]
		return msg
	}
	msg.Command = raw[pos : pos+end]
	pos = pos + end + 1

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse params
	for pos < len(raw) {
		if raw[pos] == ':' {
			// Trailing parameter (rest of message)
			msg.Trailing = raw[pos+1:]
			break
		}

		end := strings.Index(raw[pos:], " ")
		if end == -1 {
			msg.Params = append(msg.Params, raw[pos:])
			break
		}
		msg.Params = append(msg.Params, raw[pos:pos+end])
		pos = pos + end + 1

		// Skip whitespace
		for pos < len(raw) && raw[pos] == ' ' {
			pos++
		}
	}

	return msg
}

// parseTags parses IRCv3 tags from a tag string.
func parseTags(tagStr string) map[string]string {
	tags := make(map[string]string)
	if tagStr == "" {
		return tags
	}

	pairs := strings.SplitSeq(tagStr, ";")
	for pair := range pairs {
		if pair == "" {
			continue
		}
		before, after, ok := strings.Cut(pair, "=")
		if !ok {
			tags[pair] = ""
			continue
		}
		key := before
		value := unescapeTagValue(after)
		tags[key] = value
	}

	return tags
}

// unescapeTagValue unescapes IRC tag values.
func unescapeTagValue(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\\' {
			switch s[i+1] {
			case ':':
				result.WriteByte(';')
			case 's':
				result.WriteByte(' ')
			case '\\':
				result.WriteByte('\\')
			case 'r':
				result.WriteByte('\r')
			case 'n':
				result.WriteByte('\n')
			default:
				result.WriteByte(s[i+1])
			}
			i += 2
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// parseEmotes parses the emotes tag into a slice of IRCEmote.
// Format: emote_id:start-end,start-end/emote_id:start-end
func parseEmotes(emoteStr string) []IRCEmote {
	if emoteStr == "" {
		return nil
	}

	var emotes []IRCEmote
	emoteParts := strings.SplitSeq(emoteStr, "/")

	for part := range emoteParts {
		if part == "" {
			continue
		}

		before, after, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}

		emoteID := before
		positionsStr := after
		positions := strings.SplitSeq(positionsStr, ",")

		for posStr := range positions {
			before, after, ok := strings.Cut(posStr, "-")
			if !ok {
				continue
			}

			start, err1 := strconv.Atoi(before)
			end, err2 := strconv.Atoi(after)
			if err1 != nil || err2 != nil {
				continue
			}

			emotes = append(emotes, IRCEmote{
				ID:    emoteID,
				Start: start,
				End:   end,
				Count: 1,
			})
		}
	}

	return emotes
}

// parseBadges parses the badges tag into a map.
// Format: badge/version,badge/version
func parseBadges(badgeStr string) map[string]string {
	badges := make(map[string]string)
	if badgeStr == "" {
		return badges
	}

	parts := strings.SplitSeq(badgeStr, ",")
	for part := range parts {
		if part == "" {
			continue
		}
		before, after, ok := strings.Cut(part, "/")
		if !ok {
			badges[part] = ""
			continue
		}
		badges[before] = after
	}

	return badges
}

// parseTimestamp parses the tmi-sent-ts tag into a time.Time.
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}
	ms, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.UnixMilli(ms)
}

// parseBool parses a "0" or "1" string into a bool.
func parseBool(s string) bool {
	return s == "1"
}

// parseInt parses a string into an int, returning 0 on error.
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// parseChannel removes the # prefix from a channel name.
func parseChannel(s string) string {
	return strings.TrimPrefix(s, "#")
}

// parseUserFromPrefix extracts the username from a prefix.
// Format: nick!user@host
func parseUserFromPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	before, _, ok := strings.Cut(prefix, "!")
	if !ok {
		return prefix
	}
	return before
}

// parseChatMessage converts an IRCMessage into a ChatMessage.
func parseChatMessage(msg *IRCMessage) *ChatMessage {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	badges := parseBadges(msg.Tags["badges"])

	return &ChatMessage{
		ID:                     msg.Tags["id"],
		Channel:                channel,
		User:                   msg.Tags["login"],
		UserID:                 msg.Tags["user-id"],
		Message:                msg.Trailing,
		Emotes:                 parseEmotes(msg.Tags["emotes"]),
		Badges:                 badges,
		BadgeInfo:              parseBadges(msg.Tags["badge-info"]),
		Color:                  msg.Tags["color"],
		DisplayName:            msg.Tags["display-name"],
		IsMod:                  parseBool(msg.Tags["mod"]),
		IsVIP:                  badges["vip"] != "",
		IsSubscriber:           parseBool(msg.Tags["subscriber"]),
		IsBroadcaster:          badges["broadcaster"] != "",
		Bits:                   parseInt(msg.Tags["bits"]),
		FirstMessage:           parseBool(msg.Tags["first-msg"]),
		ReturningChatter:       parseBool(msg.Tags["returning-chatter"]),
		ReplyParentMsgID:       msg.Tags["reply-parent-msg-id"],
		ReplyParentUserID:      msg.Tags["reply-parent-user-id"],
		ReplyParentUserLogin:   msg.Tags["reply-parent-user-login"],
		ReplyParentDisplayName: msg.Tags["reply-parent-display-name"],
		ReplyParentMsgBody:     msg.Tags["reply-parent-msg-body"],
		Timestamp:              parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:                    msg.Raw,
	}
}

// parseUserNotice converts an IRCMessage into a UserNotice.
func parseUserNotice(msg *IRCMessage) *UserNotice {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	// Extract msg-param-* tags
	msgParams := make(map[string]string)
	for key, value := range msg.Tags {
		if after, ok := strings.CutPrefix(key, "msg-param-"); ok {
			paramName := after
			msgParams[paramName] = value
		}
	}

	return &UserNotice{
		Type:          msg.Tags["msg-id"],
		Channel:       channel,
		User:          msg.Tags["login"],
		UserID:        msg.Tags["user-id"],
		DisplayName:   msg.Tags["display-name"],
		Message:       msg.Trailing,
		SystemMessage: msg.Tags["system-msg"],
		MsgParams:     msgParams,
		Badges:        parseBadges(msg.Tags["badges"]),
		BadgeInfo:     parseBadges(msg.Tags["badge-info"]),
		Color:         msg.Tags["color"],
		Emotes:        parseEmotes(msg.Tags["emotes"]),
		Timestamp:     parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:           msg.Raw,
	}
}

// parseRoomState converts an IRCMessage into a RoomState.
func parseRoomState(msg *IRCMessage) *RoomState {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	followersOnly := -1
	if fo, ok := msg.Tags["followers-only"]; ok {
		followersOnly = parseInt(fo)
	}

	return &RoomState{
		Channel:       channel,
		EmoteOnly:     parseBool(msg.Tags["emote-only"]),
		FollowersOnly: followersOnly,
		R9K:           parseBool(msg.Tags["r9k"]),
		Slow:          parseInt(msg.Tags["slow"]),
		SubsOnly:      parseBool(msg.Tags["subs-only"]),
		RoomID:        msg.Tags["room-id"],
		Raw:           msg.Raw,
	}
}

// parseNotice converts an IRCMessage into a Notice.
func parseNotice(msg *IRCMessage) *Notice {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &Notice{
		Channel: channel,
		Message: msg.Trailing,
		MsgID:   msg.Tags["msg-id"],
		Raw:     msg.Raw,
	}
}

// parseClearChat converts an IRCMessage into a ClearChat.
func parseClearChat(msg *IRCMessage) *ClearChat {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &ClearChat{
		Channel:      channel,
		User:         msg.Trailing,
		BanDuration:  parseInt(msg.Tags["ban-duration"]),
		RoomID:       msg.Tags["room-id"],
		TargetUserID: msg.Tags["target-user-id"],
		Timestamp:    parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:          msg.Raw,
	}
}

// parseClearMessage converts an IRCMessage into a ClearMessage.
func parseClearMessage(msg *IRCMessage) *ClearMessage {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &ClearMessage{
		Channel:     channel,
		User:        msg.Tags["login"],
		Message:     msg.Trailing,
		TargetMsgID: msg.Tags["target-msg-id"],
		Timestamp:   parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:         msg.Raw,
	}
}

// parseWhisper converts an IRCMessage into a Whisper.
func parseWhisper(msg *IRCMessage) *Whisper {
	to := ""
	if len(msg.Params) > 0 {
		to = msg.Params[0]
	}

	return &Whisper{
		From:        parseUserFromPrefix(msg.Prefix),
		FromID:      msg.Tags["user-id"],
		To:          to,
		Message:     msg.Trailing,
		DisplayName: msg.Tags["display-name"],
		Color:       msg.Tags["color"],
		Badges:      parseBadges(msg.Tags["badges"]),
		Emotes:      parseEmotes(msg.Tags["emotes"]),
		MessageID:   msg.Tags["message-id"],
		ThreadID:    msg.Tags["thread-id"],
		Raw:         msg.Raw,
	}
}

// parseGlobalUserState converts an IRCMessage into a GlobalUserState.
func parseGlobalUserState(msg *IRCMessage) *GlobalUserState {
	emoteSets := []string{}
	if es := msg.Tags["emote-sets"]; es != "" {
		emoteSets = strings.Split(es, ",")
	}

	return &GlobalUserState{
		UserID:      msg.Tags["user-id"],
		DisplayName: msg.Tags["display-name"],
		Color:       msg.Tags["color"],
		Badges:      parseBadges(msg.Tags["badges"]),
		BadgeInfo:   parseBadges(msg.Tags["badge-info"]),
		EmoteSets:   emoteSets,
		Raw:         msg.Raw,
	}
}

// parseUserState converts an IRCMessage into a UserState.
func parseUserState(msg *IRCMessage) *UserState {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	emoteSets := []string{}
	if es := msg.Tags["emote-sets"]; es != "" {
		emoteSets = strings.Split(es, ",")
	}

	badges := parseBadges(msg.Tags["badges"])

	return &UserState{
		Channel:      channel,
		DisplayName:  msg.Tags["display-name"],
		Color:        msg.Tags["color"],
		Badges:       badges,
		BadgeInfo:    parseBadges(msg.Tags["badge-info"]),
		EmoteSets:    emoteSets,
		IsMod:        parseBool(msg.Tags["mod"]),
		IsSubscriber: parseBool(msg.Tags["subscriber"]),
		Raw:          msg.Raw,
	}
}
