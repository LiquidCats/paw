package metrics

import "github.com/LiquidCats/paw/watcher/internal/app/domain/entities"

type RequestToNodeCounter interface {
	Inc(chain entities.Chain)
}
