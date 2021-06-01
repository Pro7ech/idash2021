package utils


import(
	"os"
)


func FileToByteBuffer(filename string) (buffer []byte, err error){
	file, err := os.Open(filename)
	if err != nil{
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil{
		return nil, err
	}
	buffer = make([]byte, fileInfo.Size())
	if _, err = file.Read(buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}