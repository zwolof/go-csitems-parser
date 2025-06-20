package parsers

import (
	"context"
	"regexp"
	"time"

	"go-csitems-parser/models"

	"github.com/baldurstod/vdf"

	"github.com/rs/zerolog"
)

func ParseItemSets(ctx context.Context, ig *models.ItemsGame) []models.ItemSet {
	logger := zerolog.Ctx(ctx);

	start := time.Now()
	// logger.Info().Msg("Parsing item sets...")

	item_sets, err := ig.Get("item_sets")
	
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get item_sets from items_game.txt")
		return nil
	}

	var sets []models.ItemSet
	for _, s := range item_sets.GetChilds() {
		name, _ := s.GetString("name")
		set_description, _ := s.GetString("set_description")
		is_collection, _ := s.GetBool("is_collection")
		
		current := models.ItemSet{
			Key: s.Key,
			Name: name,
			SetDescription: set_description,
			IsCollection: is_collection,
			Type: models.ItemSetTypePaintKits,
		}

		// Get the items and convert them to ItemSetItem
		itemset_items, _ := s.Get("items")
		items := GetItemSetPaintKits(itemset_items)

		if len(items) == 0 {
			agents := GetItemSetAgents(itemset_items)

			if len(agents) > 0 {
				current.Agents = agents
				current.Type = models.ItemSetTypeAgents
			} else {
				continue 
			}
		} else {
			current.Items = items
		}

		// We're done here, add the current item set to the list
		sets = append(sets, current)
	}

	// Save music kits to the database
	duration := time.Since(start)
	logger.Info().Msgf("Parsed '%d' item sets in %s", len(sets), duration)

	return sets
}

func GetItemSetAgents(kv *vdf.KeyValue) []string {
	agents := make([]string, 0)

	for _, item := range kv.GetChilds() {
		agents = append(agents, item.Key)
	}

	return agents
}

func GetItemSetPaintKits(kv *vdf.KeyValue) []models.ItemSetItem {
	skins := make([]models.ItemSetItem, 0)

	// we have "[cu_tec9_asiimov]weapon_tec9" and we need to split it into "cu_tec9_asiimov" and "weapon_tec9"
	r := regexp.MustCompile(`^\[(.+?)\](.+)$`)

	for _, skin := range kv.GetChilds() {
		res := r.FindStringSubmatch(skin.Key)

		if len(res) < 3 {
			continue // skip if we can't match the pattern
		}

		paintkit_name := res[1]
		weapon_class := res[2]

		skins = append(skins, models.ItemSetItem{
			PaintKitName: paintkit_name,
			WeaponClass: weapon_class,
		})
	}

	return skins
}