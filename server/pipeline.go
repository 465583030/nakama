// Copyright 2017 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"nakama/pkg/social"
)

type pipeline struct {
	config          Config
	db              *sql.DB
	socialClient    *social.Client
	tracker         Tracker
	matchmaker      Matchmaker
	hmacSecretByte  []byte
	messageRouter   MessageRouter
	sessionRegistry *SessionRegistry
}

// NewPipeline creates a new Pipeline
func NewPipeline(config Config, db *sql.DB, socialClient *social.Client, tracker Tracker, matchmaker Matchmaker, messageRouter MessageRouter, registry *SessionRegistry) *pipeline {
	return &pipeline{
		config:          config,
		db:              db,
		socialClient:    socialClient,
		tracker:         tracker,
		matchmaker:      matchmaker,
		hmacSecretByte:  []byte(config.GetSession().EncryptionKey),
		messageRouter:   messageRouter,
		sessionRegistry: registry,
	}
}

func (p *pipeline) processRequest(logger *zap.Logger, session *session, envelope *Envelope) {
	logger.Debug(fmt.Sprintf("Received %T message", envelope.Payload))

	switch envelope.Payload.(type) {
	case *Envelope_Logout:
		// TODO Store JWT into a blacklist until remaining JWT expiry.
		p.sessionRegistry.remove(session)
		session.close()

	case *Envelope_Link:
		p.linkID(logger, session, envelope)
	case *Envelope_Unlink:
		p.unlinkID(logger, session, envelope)

	case *Envelope_SelfFetch:
		p.selfFetch(logger, session, envelope)
	case *Envelope_SelfUpdate:
		p.selfUpdate(logger, session, envelope)
	case *Envelope_UsersFetch:
		p.usersFetch(logger, session, envelope)

	case *Envelope_FriendAdd:
		p.friendAdd(logger, session, envelope)
	case *Envelope_FriendRemove:
		p.friendRemove(logger, session, envelope)
	case *Envelope_FriendBlock:
		p.friendBlock(logger, session, envelope)
	case *Envelope_FriendsList:
		p.friendsList(logger, session, envelope)

	case *Envelope_GroupCreate:
		p.groupCreate(logger, session, envelope)
	case *Envelope_GroupUpdate:
		p.groupUpdate(logger, session, envelope)
	case *Envelope_GroupRemove:
		p.groupRemove(logger, session, envelope)
	case *Envelope_GroupsFetch:
		p.groupsFetch(logger, session, envelope)
	case *Envelope_GroupsList:
		p.groupsList(logger, session, envelope)
	case *Envelope_GroupsSelfList:
		p.groupsSelfList(logger, session, envelope)
	case *Envelope_GroupUsersList:
		p.groupUsersList(logger, session, envelope)
	case *Envelope_GroupJoin:
		p.groupJoin(logger, session, envelope)
	case *Envelope_GroupLeave:
		p.groupLeave(logger, session, envelope)
	case *Envelope_GroupUserAdd:
		p.groupUserAdd(logger, session, envelope)
	case *Envelope_GroupUserKick:
		p.groupUserKick(logger, session, envelope)
	case *Envelope_GroupUserPromote:
		p.groupUserPromote(logger, session, envelope)

	case *Envelope_TopicJoin:
		p.topicJoin(logger, session, envelope)
	case *Envelope_TopicLeave:
		p.topicLeave(logger, session, envelope)
	case *Envelope_TopicMessageSend:
		p.topicMessageSend(logger, session, envelope)
	case *Envelope_TopicMessagesList:
		p.topicMessagesList(logger, session, envelope)

	case *Envelope_MatchCreate:
		p.matchCreate(logger, session, envelope)
	case *Envelope_MatchJoin:
		p.matchJoin(logger, session, envelope)
	case *Envelope_MatchLeave:
		p.matchLeave(logger, session, envelope)
	case *Envelope_MatchDataSend:
		p.matchDataSend(logger, session, envelope)

	case *Envelope_MatchmakeAdd:
		p.matchmakeAdd(logger, session, envelope)
	case *Envelope_MatchmakeRemove:
		p.matchmakeRemove(logger, session, envelope)

	case *Envelope_StorageFetch:
		p.storageFetch(logger, session, envelope)
	case *Envelope_StorageWrite:
		p.storageWrite(logger, session, envelope)
	case *Envelope_StorageRemove:
		p.storageRemove(logger, session, envelope)

	case *Envelope_LeaderboardsList:
		p.leaderboardsList(logger, session, envelope)
	case *Envelope_LeaderboardRecordWrite:
		p.leaderboardRecordWrite(logger, session, envelope)
	case *Envelope_LeaderboardRecordsFetch:
		p.leaderboardRecordsFetch(logger, session, envelope)
	case *Envelope_LeaderboardRecordsList:
		p.leaderboardRecordsList(logger, session, envelope)

	case nil:
		session.Send(ErrorMessage(envelope.CollationId, MISSING_PAYLOAD, "No payload found"))
	default:
		session.Send(ErrorMessage(envelope.CollationId, UNRECOGNIZED_PAYLOAD, "Unrecognized payload"))
	}
}

func ErrorMessageRuntimeException(collationID string, message string) *Envelope {
	return ErrorMessage(collationID, RUNTIME_EXCEPTION, message)
}

func ErrorMessageBadInput(collationID string, message string) *Envelope {
	return ErrorMessage(collationID, BAD_INPUT, message)
}

func ErrorMessage(collationID string, code Error_Code, message string) *Envelope {
	return &Envelope{
		CollationId: collationID,
		Payload: &Envelope_Error{&Error{
			Message: message,
			Code:    int32(code),
		}}}
}
