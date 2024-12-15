package checker

import (
	"goantisc/internal/log"
	"regexp"
	"strings"

	"github.com/longbridgeapp/opencc"
)

var chPtn = regexp.MustCompile(`\p{Han}`)

func NewSimplifiedChineseChecker(occ *opencc.OpenCC) *SimplifiedChineseChecker {
	return &SimplifiedChineseChecker{
		dict: occ,
	}
}

type SimplifiedChineseChecker struct {
	dict *opencc.OpenCC
}

func (si *SimplifiedChineseChecker) FilterForbiddenWords(sentence string) (diffWord []string, originNum int) {
	allChStrs := chPtn.FindAllString(sentence, -1)

	fullChSentence := strings.Join(allChStrs, "")
	if fullChSentence == "" {
		return []string{}, len(allChStrs)
	}

	originStr := fullChSentence
	tcStr, err := si.dict.Convert(sentence)
	if err != nil {
		log.Logger().Errorf("[OpenCC] Convert failed, err: %v, str: %v", sentence)
		return []string{}, len(allChStrs)
	}

	// compare origin & tc string
	var differences []string
	runesOrigin := []rune(originStr)
	runesTC := []rune(tcStr)

	for i := 0; i < len(runesOrigin) && i < len(runesTC); i++ {
		if runesOrigin[i] != runesTC[i] {
			differences = append(differences, string(runesOrigin[i]))
		}
	}

	if len(runesOrigin) > len(runesTC) {
		for i := len(runesTC); i < len(runesOrigin); i++ {
			differences = append(differences, string(runesOrigin[i]))
		}
	} else if len(runesTC) > len(runesOrigin) {
		for i := len(runesOrigin); i < len(runesTC); i++ {
			differences = append(differences, string(runesTC[i]))
		}
	}
	return differences, len(allChStrs)
}
