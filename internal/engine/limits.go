package engine

const (
	// MaxBatchesPerFile is the maximum number of batches ("lotes") a
	// single CNAB 240 file may contain, per the FEBRABAN standard.
	MaxBatchesPerFile = 70
	// MaxMovementsPerBatch is the maximum number of movements (payments)
	// a single batch may contain, per the FEBRABAN standard.
	MaxMovementsPerBatch = 10000

	// LimitBatchesPerFile identifies the batches-per-file limit in a
	// LimitError.
	LimitBatchesPerFile = "batches_per_file"
	// LimitMovementsPerBatch identifies the movements-per-batch limit in
	// a LimitError.
	LimitMovementsPerBatch = "movements_per_batch"
)
