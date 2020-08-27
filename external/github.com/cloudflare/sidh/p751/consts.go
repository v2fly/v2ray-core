package p751

import (
	. "github.com/v2fly/v2ray-core/external/github.com/cloudflare/sidh/internal/isogeny"
	cpu "github.com/v2fly/v2ray-core/external/github.com/cloudflare/sidh/internal/utils"
)

const (
	// SIDH public key byte size
	P751_PublicKeySize = 564
	// SIDH shared secret byte size.
	P751_SharedSecretSize = 188
	// Max size of secret key for 2-torsion group, corresponds to 2^e2
	P751_SecretBitLenA = 372
	// Size of secret key for 3-torsion group, corresponds to floor(log_2(3^e3))
	P751_SecretBitLenB = 378
	// P751 bytelen ceil(751/8)
	P751_Bytelen = 94
	// Size of a compuatation strategy for 2-torsion group
	strategySizeA = 185
	// Size of a compuatation strategy for 3-torsion group
	strategySizeB = 238
	// Number of 64-bit limbs used to store Fp element
	NumWords = 12
)

// CPU Capabilities. Those flags are referred by assembly code. According to
// https://github.com/golang/go/issues/28230, variables referred from the
// assembly must be in the same package.
// We declare them variables not constants in order to facilitate testing.
var (
	// Signals support for MULX which is in BMI2
	HasBMI2 = cpu.X86.HasBMI2
	// Signals support for ADX and BMI2
	HasADXandBMI2 = cpu.X86.HasBMI2 && cpu.X86.HasADX
)

// The x-coordinate of PA
var P751_affine_PA = Fp2Element{
	A: FpElement{
		0xC2FC08CEAB50AD8B, 0x1D7D710F55E457B1, 0xE8738D92953DCD6E,
		0xBAA7EBEE8A3418AA, 0xC9A288345F03F46F, 0xC8D18D167CFE2616,
		0x02043761F6B1C045, 0xAA1975E13180E7E9, 0x9E13D3FDC6690DE6,
		0x3A024640A3A3BB4F, 0x4E5AD44E6ACBBDAE, 0x0000544BEB561DAD,
	},
	B: FpElement{
		0xE6CC41D21582E411, 0x07C2ECB7C5DF400A, 0xE8E34B521432AEC4,
		0x50761E2AB085167D, 0x032CFBCAA6094B3C, 0x6C522F5FDF9DDD71,
		0x1319217DC3A1887D, 0xDC4FB25803353A86, 0x362C8D7B63A6AB09,
		0x39DCDFBCE47EA488, 0x4C27C99A2C28D409, 0x00003CB0075527C4,
	},
}

// The x-coordinate of QA
var P751_affine_QA = Fp2Element{
	A: FpElement{
		0xD56FE52627914862, 0x1FAD60DC96B5BAEA, 0x01E137D0BF07AB91,
		0x404D3E9252161964, 0x3C5385E4CD09A337, 0x4476426769E4AF73,
		0x9790C6DB989DFE33, 0xE06E1C04D2AA8B5E, 0x38C08185EDEA73B9,
		0xAA41F678A4396CA6, 0x92B9259B2229E9A0, 0x00002F9326818BE0,
	},
	B: FpElement{
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
	},
}

// The x-coordinate of RA = PA-QA
var P751_affine_RA = Fp2Element{
	A: FpElement{
		0x0BB84441DFFD19B3, 0x84B4DEA99B48C18E, 0x692DE648AD313805,
		0xE6D72761B6DFAEE0, 0x223975C672C3058D, 0xA0FDE0C3CBA26FDC,
		0xA5326132A922A3CA, 0xCA5E7F5D5EA96FA4, 0x127C7EFE33FFA8C6,
		0x4749B1567E2A23C4, 0x2B7DF5B4AF413BFA, 0x0000656595B9623C,
	},
	B: FpElement{
		0xED78C17F1EC71BE8, 0xF824D6DF753859B1, 0x33A10839B2A8529F,
		0xFC03E9E25FDEA796, 0xC4708A8054DF1762, 0x4034F2EC034C6467,
		0xABFB70FBF06ECC79, 0xDABE96636EC108B7, 0x49CBCFB090605FD3,
		0x20B89711819A45A7, 0xFB8E1590B2B0F63E, 0x0000556A5F964AB2,
	},
}

// The x-coordinate of PB
var P751_affine_PB = Fp2Element{
	A: FpElement{
		0xCFB6D71EF867AB0B, 0x4A5FDD76E9A45C76, 0x38B1EE69194B1F03,
		0xF6E7B18A7761F3F0, 0xFCF01A486A52C84C, 0xCBE2F63F5AA75466,
		0x6487BCE837B5E4D6, 0x7747F5A8C622E9B8, 0x4CBFE1E4EE6AEBBA,
		0x8A8616A13FA91512, 0x53DB980E1579E0A5, 0x000058FEBFF3BE69,
	},
	B: FpElement{
		0xA492034E7C075CC3, 0x677BAF00B04AA430, 0x3AAE0C9A755C94C8,
		0x1DC4B064E9EBB08B, 0x3684EDD04E826C66, 0x9BAA6CB661F01B22,
		0x20285A00AD2EFE35, 0xDCE95ABD0497065F, 0x16C7FBB3778E3794,
		0x26B3AC29CEF25AAF, 0xFB3C28A31A30AC1D, 0x000046ED190624EE,
	},
}

// The x-coordinate of QB
var P751_affine_QB = Fp2Element{
	A: FpElement{
		0xF1A8C9ED7B96C4AB, 0x299429DA5178486E, 0xEF4926F20CD5C2F4,
		0x683B2E2858B4716A, 0xDDA2FBCC3CAC3EEB, 0xEC055F9F3A600460,
		0xD5A5A17A58C3848B, 0x4652D836F42EAED5, 0x2F2E71ED78B3A3B3,
		0xA771C057180ADD1D, 0xC780A5D2D835F512, 0x0000114EA3B55AC1,
	},
	B: FpElement{
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
	},
}

// The x-coordinate of RB = PB - QB
var P751_affine_RB = Fp2Element{
	A: FpElement{
		0x1C0D6733769D0F31, 0xF084C3086E2659D1, 0xE23D5DA27BCBD133,
		0xF38EC9A8D5864025, 0x6426DC781B3B645B, 0x4B24E8E3C9FB03EE,
		0x6432792F9D2CEA30, 0x7CC8E8B1AE76E857, 0x7F32BFB626BB8963,
		0xB9F05995B48D7B74, 0x4D71200A7D67E042, 0x0000228457AF0637,
	},
	B: FpElement{
		0x4AE37E7D8F72BD95, 0xDD2D504B3E993488, 0x5D14E7FA1ECB3C3E,
		0x127610CEB75D6350, 0x255B4B4CAC446B11, 0x9EA12336C1F70CAF,
		0x79FA68A2147BC2F8, 0x11E895CFDADBBC49, 0xE4B9D3C4D6356C18,
		0x44B25856A67F951C, 0x5851541F61308D0B, 0x00002FFD994F7E4C,
	},
}

// 2-torsion group computation strategy
var P751_AliceIsogenyStrategy = [strategySizeA]uint32{
	0x50, 0x30, 0x1B, 0x0F, 0x08, 0x04, 0x02, 0x01, 0x01, 0x02,
	0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x07,
	0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x03, 0x02, 0x01,
	0x01, 0x01, 0x01, 0x0C, 0x07, 0x04, 0x02, 0x01, 0x01, 0x02,
	0x01, 0x01, 0x03, 0x02, 0x01, 0x01, 0x01, 0x01, 0x05, 0x03,
	0x02, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x01, 0x15,
	0x0C, 0x07, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x03,
	0x02, 0x01, 0x01, 0x01, 0x01, 0x05, 0x03, 0x02, 0x01, 0x01,
	0x01, 0x01, 0x02, 0x01, 0x01, 0x01, 0x09, 0x05, 0x03, 0x02,
	0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x01, 0x04, 0x02,
	0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x21, 0x14, 0x0C, 0x07,
	0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x03, 0x02, 0x01,
	0x01, 0x01, 0x01, 0x05, 0x03, 0x02, 0x01, 0x01, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x01, 0x08, 0x05, 0x03, 0x02, 0x01, 0x01,
	0x01, 0x01, 0x02, 0x01, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x10, 0x08, 0x04, 0x02, 0x01, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01,
	0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02,
	0x01, 0x01, 0x02, 0x01, 0x01}

// 3-torsion group computation strategy
var P751_BobIsogenyStrategy = [strategySizeB]uint32{
	0x70, 0x3F, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01, 0x01, 0x02,
	0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x08,
	0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02, 0x01,
	0x01, 0x02, 0x01, 0x01, 0x10, 0x08, 0x04, 0x02, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01,
	0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02,
	0x01, 0x01, 0x02, 0x01, 0x01, 0x1F, 0x10, 0x08, 0x04, 0x02,
	0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02,
	0x01, 0x01, 0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01,
	0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x0F, 0x08, 0x04,
	0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x07, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01,
	0x01, 0x03, 0x02, 0x01, 0x01, 0x01, 0x01, 0x31, 0x1F, 0x10,
	0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04, 0x02,
	0x01, 0x01, 0x02, 0x01, 0x01, 0x08, 0x04, 0x02, 0x01, 0x01,
	0x02, 0x01, 0x01, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01,
	0x0F, 0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x04,
	0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x07, 0x04, 0x02, 0x01,
	0x01, 0x02, 0x01, 0x01, 0x03, 0x02, 0x01, 0x01, 0x01, 0x01,
	0x15, 0x0C, 0x08, 0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01,
	0x04, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0x05, 0x03, 0x02,
	0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x01, 0x09, 0x05,
	0x03, 0x02, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x01,
	0x04, 0x02, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01}

// Used internally by this package. Not consts as Go doesn't allow arrays to be consts
// -------------------------------

// p751
var p751 = FpElement{
	0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
	0xffffffffffffffff, 0xffffffffffffffff, 0xeeafffffffffffff,
	0xe3ec968549f878a8, 0xda959b1a13f7cc76, 0x084e9867d6ebe876,
	0x8562b5045cb25748, 0x0e12909f97badc66, 0x00006fe5d541f71c}

// 2*p751
var p751x2 = FpElement{
	0xFFFFFFFFFFFFFFFE, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF,
	0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xDD5FFFFFFFFFFFFF,
	0xC7D92D0A93F0F151, 0xB52B363427EF98ED, 0x109D30CFADD7D0ED,
	0x0AC56A08B964AE90, 0x1C25213F2F75B8CD, 0x0000DFCBAA83EE38}

// p751 + 1
var p751p1 = FpElement{
	0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
	0x0000000000000000, 0x0000000000000000, 0xeeb0000000000000,
	0xe3ec968549f878a8, 0xda959b1a13f7cc76, 0x084e9867d6ebe876,
	0x8562b5045cb25748, 0x0e12909f97badc66, 0x00006fe5d541f71c}

// R^2 = (2^768)^2 mod p
var p751R2 = FpElement{
	2535603850726686808, 15780896088201250090, 6788776303855402382,
	17585428585582356230, 5274503137951975249, 2266259624764636289,
	11695651972693921304, 13072885652150159301, 4908312795585420432,
	6229583484603254826, 488927695601805643, 72213483953973}

// 1*R mod p
var P751_OneFp2 = Fp2Element{
	A: FpElement{
		0x249ad, 0x0, 0x0, 0x0, 0x0, 0x8310000000000000, 0x5527b1e4375c6c66, 0x697797bf3f4f24d0, 0xc89db7b2ac5c4e2e, 0x4ca4b439d2076956, 0x10f7926c7512c7e9, 0x2d5b24bce5e2},
}

// 1/2 * R mod p
var P751_HalfFp2 = Fp2Element{
	A: FpElement{
		0x00000000000124D6, 0x0000000000000000, 0x0000000000000000,
		0x0000000000000000, 0x0000000000000000, 0xB8E0000000000000,
		0x9C8A2434C0AA7287, 0xA206996CA9A378A3, 0x6876280D41A41B52,
		0xE903B49F175CE04F, 0x0F8511860666D227, 0x00004EA07CFF6E7F},
}
