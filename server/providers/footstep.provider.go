package providers

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"horizon.com/server/server/model"
)

func (p *Providers) UserFootstep(
	c echo.Context,
	organizationID, branchID string,
	module, activity, description string,
) {
	go func() {
		userOrg, err := p.CurrentUserOrganization(c, organizationID, branchID)
		if err != nil {
			return
		}
		fs := &model.Footstep{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
			CreatedByID:    userOrg.UserID,
			UpdatedByID:    userOrg.UserID,
			UserID:         &userOrg.UserID,
			Module:         module,
			Activity:       activity,
			Description:    description,
			Timestamp:      time.Now(),
			IPAddress:      c.RealIP(),
			UserAgent:      c.Request().UserAgent(),
			Referer:        c.Request().Referer(),
			AcceptLanguage: c.Request().Header.Get("Accept-Language"),
		}
		if err := p.footstep.Manager.Create(fs); err != nil {
			_ = eris.Wrap(err, "UserFootstep: Create failed")
		}
	}()
}
