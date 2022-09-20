package fscommon

type TempfilenameGenerator func(fullpath string) string

func TempfilenameGeneratorBuilderSimpleNew(suffix string) TempfilenameGenerator {
	return func(fullpath string) string {
		return fullpath + suffix
	}
}

var TempfilenameGeneratorSimpleDefault TempfilenameGenerator = TempfilenameGeneratorBuilderSimpleNew(".tmp")
