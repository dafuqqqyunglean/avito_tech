package mapper

import "github.com/dafuqqqyunglean/avito_tech/domain"

func FromReqToResp(req domain.TeamRequest) domain.TeamResponse {
	return domain.TeamResponse{
		Team: req,
	}
}
