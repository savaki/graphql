// generated by stringer -type=itemType; DO NOT EDIT

package parse

import "fmt"

const _itemType_name = "itemErroritemBoolitemEOFitemNameitemQueryitemLeftCurlyitemRightCurlyitemLeftParenitemRightParenitemNumberitemColonitemCommaitemKeyworditemDotitemNilitemSpaceitemTextitemStringitemRawStringitemIdentifieritemCharitemCharConstantitemComplex"

var _itemType_index = [...]uint8{0, 9, 17, 24, 32, 41, 54, 68, 81, 95, 105, 114, 123, 134, 141, 148, 157, 165, 175, 188, 202, 210, 226, 237}

func (i itemType) String() string {
	if i < 0 || i+1 >= itemType(len(_itemType_index)) {
		return fmt.Sprintf("itemType(%d)", i)
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
