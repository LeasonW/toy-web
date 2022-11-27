package file

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_Demo(t *testing.T) {
	fmt.Println(os.Getwd())

	f, err := os.Open("../testdata/filetest/myfile")
	require.NoError(t, err)

	data := make([]byte, 64)
	n, err := f.Read(data)
	require.NoError(t, err)
	println(n)
	f.Close()

	// 追加写
	f, err = os.OpenFile("../testdata/filetest/myfile", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	require.NoError(t, err)

	n, err = f.WriteString("hello world")
	require.NoError(t, err)
	println(n)
	f.Close()

	f, err = os.Create("../testdata/filetest/myfile_copy")
	require.NoError(t, err)
	n, err = f.WriteString("hello world")
	require.NoError(t, err)
	println(n)
}
