package remotepkg

import "encoding/binary"

type AliasType = binary.ByteOrder

const AliasConst = binary.MaxVarintLen64

var AliasVar = binary.LittleEndian
