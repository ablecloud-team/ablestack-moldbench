# Ablestack - moldbench
Mold의 성능과 효율성을 평가하도록 설계된 도구입니다.

## 실행방법
1. golang 설치 및 환경변수설정
```bash
go version
```
2. 패키지와 관련 종속성 다운로드 및 설치
```bash
go get
```
3. 빌드
```bash
go build . 또는 go build moldbench.go
```
4. 벤치마킹 실행
```bash
go run . -simpletest 또는 ./moldbench -simpletest
```

## 오류해결
1. go 명령어 실행시 오류
```bash
zsh: command not found: go
```
> 환경변수 설정

2. GO111MODULE 설정 문제
```bash
go: modules disabled by GO111MODULE=off; see 'go help modules'
```
> export GO111MODULE="on"

3. cannot find package
```bash
moldbench.go:34:2: cannot find package "moldbench/apirunner" in any of:/usr/local/go/src/moldbench/apirunner (from $GOROOT)/Users/yusujeong/go/src moldbench/apirunner (from $GOPATH)
```
> go get

4. go mod 패키지 변경 및 추가
> go.mod 및 go.sum 파일삭제 후 재생성

> rm -rf go.mod go.sum

> go mod tidy
