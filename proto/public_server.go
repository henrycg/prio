package proto

import (
	"log"

	"github.com/henrycg/prio/config"
)

func HashToServer(cfg *config.Config, uuid Uuid) int {
	return int(uuid[0]) % cfg.NumServers()
}

func NewPublicServer(server *Server) *PublicServer {
	p := new(PublicServer)
	p.server = server
	return p
}

func (s *PublicServer) Upload(args *UploadArgs, reply *UploadReply) error {
	dstServer := HashToServer(s.server.cfg, args.PublicKey)
	if dstServer == s.server.ServerIdx {
		s.server.toProcess <- args
	} else {
		log.Printf("Ignoring request send to wrong server")
	}

	return nil
}
