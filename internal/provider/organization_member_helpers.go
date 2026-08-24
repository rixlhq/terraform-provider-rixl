package provider

import (
	"context"

	"github.com/rixlhq/rixl-go/sdk/memberships"
)

func (r *organizationMemberResource) findMember(ctx context.Context, orgID, userID string) (any, error) {
	params := &memberships.ListOrganizationMembersParams{
		Limit: new(int32(100)),
	}
	if userID != "" {
		params.UserUserId = new(userID)
	}

	listResp, err := r.client.Memberships.ListOrganizationMembers(ctx, orgID, params, nil)
	if err != nil {
		return nil, err
	}

	for i := range listResp.Members {
		member := listResp.Members[i]
		memberMap, err := responseToMap(member)
		if err != nil {
			continue
		}
		if uid, ok := memberMap["user_id"].(string); ok && uid == userID {
			return member, nil
		}
	}
	return nil, nil
}

func (r *organizationMemberResource) applyState(ctx context.Context, orgID, userID, desiredState string) error {
	switch desiredState {
	case "MEMBERSHIP_STATE_SUSPENDED":
		_, err := r.client.Memberships.SuspendMember(ctx, orgID, userID, nil)
		return err
	case "MEMBERSHIP_STATE_ACTIVE":
		_, err := r.client.Memberships.ReactivateMember(ctx, orgID, userID, nil)
		return err
	default:
		return nil
	}
}
