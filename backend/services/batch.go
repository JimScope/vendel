package services

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// processInBatches drains an eligible record set in bounded batches. It queries
// up to CronDrainBatchSize records at a time and invokes process on each,
// repeating until the eligible set is exhausted or CronMaxDrainBatches is
// reached — a hard cap that guarantees the loop is always bounded, never a
// silent 50-row cut-off (REL-4) nor an unbounded scan.
//
// process reports whether the record LEFT the eligible set: return true when
// the record no longer matches the filter (e.g. its status changed), and false
// when it stays eligible (skipped this pass, save failed, etc.). The query
// offset advances by the number of still-eligible records so those are not
// re-fetched within the same pass — otherwise permanent skips at the front of
// the set would starve the batch budget and progress would stall.
//
// The filter must impose a stable total order via sort so pagination is
// consistent across queries; callers that mutate records out of the set rely on
// still-eligible records keeping their relative position at the front.
func processInBatches(
	app core.App,
	collection, filter, sort string,
	params dbx.Params,
	process func(*core.Record) bool,
) error {
	offset := 0
	for batch := 0; batch < CronMaxDrainBatches; batch++ {
		records, err := app.FindRecordsByFilter(collection, filter, sort, CronDrainBatchSize, offset, params)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		for _, r := range records {
			if !process(r) {
				offset++ // stays eligible — skip past it on the next query
			}
		}
		if len(records) < CronDrainBatchSize {
			return nil
		}
	}
	return nil
}
