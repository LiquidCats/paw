package base58

import "math/big"

var bigRadix = [...]*big.Int{
	big.NewInt(0),                                          // nolint:mnd
	big.NewInt(58),                                         // nolint:mnd
	big.NewInt(58 * 58),                                    // nolint:mnd
	big.NewInt(58 * 58 * 58),                               // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58),                          // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58 * 58),                     // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58),                // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58),           // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58),      // nolint:mnd
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58), // nolint:mnd
	bigRadix10,
}

var bigRadix10 = big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58) // nolint:mnd
