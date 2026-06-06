# fileoptimizer

클라이언트 요구사항
- 현재 시스템의 정보를 스캔 및 수집해 매니저 서버(https 서버)에 암호화, 압축해서 전송하는 agent
- 수집할 정보는 아래와 같음
  - config에 있는 서버 내의 특정 경로 하위의 하위 디렉토리까지 전체 다 스캔하여 각 파일을 해쉬로 중복 구분함
  - 시스템 정보 (os, 고유식별번호, 호스트 이름, 시스템 성능, 기종 등)
  - 네트워크 정보 (공인 ip, private ip, 네트워크 인터페이스, 주변 네트워크 장비 목록, 주변 와이파이 탐색 목록 및 이력, 블루투스 탐색 목록 및 이력, 위치정보)
  - 디렉토리 전체 구조
- config는 서버와 동기화되어야 함, config의 내용으로 제어하고 수집할 범위를 지정함
- 필요한 권한은 실행 시 요구
- 서버에 로그인 인증에 대한 요청을 보내야됨

프로그램 구조
- 정보 수집 모듈
  - 수집할 정보에 대해서 권한 및 범위(config) 체크
  - 시스템 정보 수집
  - 네트워크 정보 수집
  - 파일 정보 수집
  - 개인 정보 수집 (위치, 계정 이름 등)
- 시스템 제어
  - 파일 이동, 삭제, 생성 -> 동일한 파일임을 보장해야함
- 인증 모듈
  - 서버와 인증정보 주고받기
  - 인증 확인, 인증 정보 암복호화 보관 등
- 암복호화, 압축 모듈
  - 수집된 정보, 서버와 주고받는 정보들에 대한 암복호화
  - 송수신시 데이터 압축(가능하다면)

서버 요구사항
- 서버에서 아래 클라이언트를 브라우저를 통해 내려받을 수 있어야 하고 웹에서 설정한 config 정보도 같이 포함되어야 함
- 간단한 JWT 인증 구현
- 웹에서 인증 후 해당 유저에 대해 수집된 정보, 제어할 정보 등을 보여줌
- react와 go backend(비동기 처리 원활한)


go routine으로 프로그램 실행과 동시에 실행될 작업
- 네트워크 관련 정보 수집
  - 현재 pc의 인터페이스
  - 인터페이스 중 ipv4 아이피를 할당받은 인터페이스가 있다면 해당 아이피 대역에 대한 네트워크 스캔
    - 22, 80 등 기본 포트들이 활성화된 ip
    - 해당 서버의 os와 가동 중인 서비스(추측도 포함)
  - 현재 클라이언트에서 사용 중인(활성화된) 포트
  - 연결된 외부 장비 목록(블루투스, 와이파이)
- 시스템 관련 정보 수집
  - 성능, 운영체제, 버전, 기타 시스템 파일 정보
- 서버와 연결 시도 (네트워크 연결)
  - 서버에서 다운로드 받을 때 각 클라이언트별 고유 코드 발급

사용자 요청 시 실행될 작업
- 로그인 시도
- 파일 정리
- 디렉토리 스캔 및 중복 파일 검사


---

## 빌드 및 배포

> 모든 명령은 프로젝트 루트에서 실행합니다.

### Makefile (권장)

```bash
make all          # macOS(Universal) + Linux amd64/arm64 + Windows 동시 빌드
                  # Linux 빌드는 Docker 필요
                  # → dist/agents/ 에 바이너리 생성

make local        # macOS + Windows 만 (Docker 없이 빌드 가능)

make darwin-universal   # macOS Universal Binary 만
make linux-amd64        # Linux amd64 (Docker 필요)
make linux-arm64        # Linux arm64 (Docker 필요)
make windows            # Windows amd64 만

make deploy       # make all 후 docker compose up -d --build
make clean        # dist/ 디렉터리 삭제
```

> **플랫폼별 CGO 요구사항**
>
> | 플랫폼 | CGO | 비고 |
> |--------|-----|------|
> | macOS (arm64/amd64) | ✅ 필요 | Cocoa 바인딩 |
> | Linux amd64/arm64 | ✅ 필요 | GTK3 바인딩 — **Docker 사용** |
> | Windows amd64 | ❌ 불필요 | 순수 Go 구현, macOS에서 크로스 컴파일 가능 |

빌드 결과물 위치:

```
dist/
└── agents/
    ├── fileoptimizer-agent-darwin          ← macOS Universal
    ├── fileoptimizer-agent-linux-amd64
    ├── fileoptimizer-agent-linux-arm64
    └── fileoptimizer-agent-windows-amd64.exe
```

> **배포 플로우 (권장)**
> ```sh
> # 에이전트 코드 수정 후
> make deploy
>
> # 단계별로 실행하려면
> make all
> docker compose up -d --build
> ```

nginx가 `dist/agents/` 를 `/agents/` 경로로 서빙하므로, 웹 대시보드 → **Download Agent** 버튼이 자동으로 연결됩니다.

---

### Agent 수동 빌드 (Makefile 없이)

Agent는 `getlantern/systray` 기반 tray 앱입니다.
- **macOS / Linux**: Cocoa(macOS) 또는 GTK(Linux) 바인딩을 사용하므로 CGO가 반드시 필요합니다.
- **Windows**: Win32 API를 순수 Go로 구현하므로 `CGO_ENABLED=0` 크로스 컴파일이 가능합니다.

| 플랫폼 | 아키텍처 | 명령어 | 출력 파일 |
|--------|----------|--------|-----------|
| **macOS** (Apple Silicon) | arm64 | `GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o fileoptimizer-agent ./agent/` | `fileoptimizer-agent` |
| **macOS** (Intel) | amd64 | `CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o fileoptimizer-agent ./agent/` | `fileoptimizer-agent` |
| **macOS** (Universal Binary) | arm64 + amd64 | 아래 참고 | `fileoptimizer-agent` |
| **Linux** | amd64 / arm64 | 아래 참고 (네이티브 빌드 필요) | `fileoptimizer-agent` |
| **Windows** | amd64 | `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o fileoptimizer-agent.exe ./agent/` | `fileoptimizer-agent.exe` |

#### macOS Universal Binary (arm64 + amd64 합치기)

> Apple Silicon(arm64) macOS에서 실행합니다.

```bash
# arm64는 네이티브 빌드
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o fileoptimizer-agent-arm64 ./agent/

# amd64는 크로스 아키텍처 → CGO_ENABLED=1 명시 필요
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o fileoptimizer-agent-amd64 ./agent/

lipo -create -output fileoptimizer-agent fileoptimizer-agent-arm64 fileoptimizer-agent-amd64
rm fileoptimizer-agent-arm64 fileoptimizer-agent-amd64
```

#### Linux 빌드 주의사항

Linux용 systray는 GTK3 + AppIndicator 라이브러리가 필요해 **macOS에서 크로스 컴파일 불가**합니다.
Linux 머신 또는 GTK 헤더가 포함된 Docker 환경에서 네이티브로 빌드해야 합니다.

```bash
# Linux 머신에서 직접 실행 (빌드 전 의존성 설치 필요)
# Ubuntu/Debian: sudo apt install libgtk-3-dev libappindicator3-dev
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o fileoptimizer-agent ./agent/
```

#### macOS에서 한 번에 빌드 (Windows 포함)

```bash
mkdir -p dist

# macOS arm64 (네이티브)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" \
  -o dist/fileoptimizer-agent-darwin-arm64 ./agent/

# macOS amd64 (크로스 아키텍처, CGO_ENABLED=1 명시)
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" \
  -o dist/fileoptimizer-agent-darwin-amd64 ./agent/

# Windows amd64 (CGO 불필요, 크로스 컴파일 가능)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" \
  -o dist/fileoptimizer-agent-windows-amd64.exe ./agent/

# Linux는 Linux 머신에서 별도 빌드 (위 Linux 빌드 주의사항 참고)
```

### Server 빌드

```bash
# 현재 플랫폼 (로컬 실행)
go build -ldflags="-s -w" -o fileoptimizer-server ./server/

# Linux amd64 (Docker / 서버 배포용)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o fileoptimizer-server-linux ./server/
```

### ldflags 설명

| 플래그 | 설명 |
|--------|------|
| `-s` | 심볼 테이블 제거 → 바이너리 크기 감소 |
| `-w` | DWARF 디버그 정보 제거 → 바이너리 크기 감소 |
| `-H windowsgui` | Windows에서 콘솔 창 숨김 (systray 앱 필수) |
| `CGO_ENABLED=0` | 순수 Go 빌드 — Windows 크로스 컴파일 시 사용 |
| `CGO_ENABLED=1` | CGO 활성화 — macOS 크로스 아키텍처 빌드 시 명시 필요 |

---

### 우분투 goenv, go 설치

- .bashrc
```bash
export GOENV_ROOT="$HOME/.goenv"
export PATH="$GOENV_ROOT/bin:$PATH"
eval "$(goenv init -)"
```

- install
```bash
git clone https://github.com/syndbg/goenv.git ~/.goenv

# .bashrc 추가

source ~/.bashrc  # 또는 source ~/.zshrc

goenv install -l  # 설치 가능한 Go 버전 목록 확인
goenv install 1.22.2  # 원하는 버전 설치

goenv global 1.22.2 # goenv local 1.22.2
goenv version

goenv uninstall 1.22.2
```


- 서버 연결
  - 인증(정보가 있는지 확인)
    - 최초 연결 시 크레덴셜 생성
    - post 서버/agent/agnet-id (크레덴셜 포함해서 정보 전송)
    - 크레덴셜 인증 후 agent_id 별로 정보 수집해서 보관
    - 보관된 정보를 웹에서 보여주고, 사용자가 기록해둔 정보를 받아옴 
      - get 서버/agent/agent-id/config

- 루틴 1 : 적당한 주기로 (하루 1~2회?)
  - config 동기화
  - 수집 후 전송
- 루틴 2 : 좀 더 자주 (30분 - 1시간에 1번?)
  - 서버 명령 대기 (실시간인지, 동기화 시에만 인지(동기화 시에만 이면 서버 부담이 적음, 에이전트가 대기하고 있지 않아도 됨))


[hash : user-name;agent-code;path ... ] 형태로 서버에 전달 -> 서버는 동일한 해쉬의 리스트에 해당 문자열을 추가

웹에서 로그인 -> agent 다운로드(이 시점에서 agent 코드생성) -> 파일 스캔 후 위 형태로 hash 생성