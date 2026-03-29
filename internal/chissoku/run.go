// Package chissoku は chissoku 子プロセスの起動と stdout 読み取りを担当する。
package chissoku

import (
	"context"
	"fmt"
	"io"
)

// RunOptions は milestone 1 では最小限のフィールドだけ用意してある。
// 後の milestone で設定ファイルに合わせて足していく。
type RunOptions struct {
	// BinPath は chissoku 実行ファイルのパス。
	BinPath string
	// Device はシリアルデバイス（例: /dev/ttyACM0）。chissoku に渡す引数は実装者が README / chissoku --help で確認すること。
	Device string
}

// ReadOneLine は chissoku を起動し、stdout から最初の 1 行を読んで返す（milestone 1 用の最小 API）。
//
// TODO(milestone-1): os/exec の CommandContext で子プロセスを起動する。
// - 引数例: デバイス指定が必要なら chissoku の実際の CLI に合わせる（--device など）。
// - Stdout をパイプし、bufio.NewScanner や ReadLine で 1 行読む。
// - プロセスがすぐ終了しないよう、chissoku は継続出力するので「1 行読めたら」プロセスを切ってよいかは chissoku の挙動に合わせて判断する。
func ReadOneLine(ctx context.Context, opt RunOptions) ([]byte, error) {
	_ = ctx
	_ = opt
	return nil, fmt.Errorf("TODO(milestone-1): ReadOneLine を実装する")
}

// StdoutReader は milestone 以降で「継続的に stdout を読む」ために使う予定のプレースホルダ。
type StdoutReader struct {
	r io.Reader
}

// NewStdoutReader は TODO(milestone-2 以降): 長時間動かす chissoku から行を読み続けるときに使う。
func NewStdoutReader(r io.Reader) *StdoutReader {
	return &StdoutReader{r: r}
}
