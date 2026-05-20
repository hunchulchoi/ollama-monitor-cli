# 📊 ntcharts 기반 반응형 차트 통합 설계 (Design Spec)

- **작성일**: 2026-05-20
- **상태**: 승인됨 (Approved)
- **대상 프로젝트**: Ollama Monitoring CLI

## 1. 개요 (Goal)
현재 유니코드 블록 기반의 단순 스파크라인 그래프를 `ntcharts` 라이브러리를 사용한 고성능 라인 차트로 교체합니다. 터미널 높이에 따라 상세한 차트와 컴팩트한 스파크라인을 자동으로 전환하는 **반응형 대시보드**를 구현하는 것이 핵심 목표입니다.

## 2. 사용자 요구사항 (Requirements)
- 모든 그래프 섹션(CPU, Memory, Latency, TPS)에 `ntcharts` 적용.
- 상세 라인 차트 스타일(X/Y축, 눈금 포함) 선호.
- 터미널 크기에 따른 반응형 레이아웃 (3단계 전환).

## 3. 상세 설계 (Technical Specification)

### 3.1 아키텍처 변경
- `internal/tui/model.go`의 `Model` 구조체에 `ntcharts.LineChart` 컴포넌트들을 임베딩합니다.
- 데이터 수집 로직(internal/ollama)은 유지하되, TUI `Update` 루프에서 데이터를 각 차트 모델로 전달하는 브릿지 로직을 추가합니다.

### 3.2 반응형 로직 (Adaptive Layout)
터미널 높이(`m.height`)에 따라 렌더링 모드를 결정합니다.

| 모드 | 임계값 (Height) | 구성 요소 |
| :--- | :--- | :--- |
| **Full Mode** | >= 40 lines | 모든 지표를 독립적인 상세 차트(8-10줄 높이)로 표시 |
| **Normal Mode** | 25 ~ 39 lines | 주요 지표만 상세 차트로 표시, 나머지는 스파크라인 |
| **Compact Mode** | < 25 lines | 모든 지표를 기존의 1줄 스파크라인으로 표시 |

### 3.3 차트 스타일링 (Styling)
- **Library**: `github.com/NimbleMarkets/ntcharts`
- **Colors**: 
  - CPU: Spring Green (Color 42)
  - Memory: Deep Sky Blue (Color 39)
  - Latency: Orange (Color 214)
- **Grid**: 옅은 점선 또는 실선 그리드 적용 (ntcharts 옵션 활용)

### 3.4 데이터 관리
- **Data Points**: 차트당 최대 100개 포인트 유지.
- **Refresh Rate**: 2초 주기의 `TickMsg`와 동기화.

## 4. 구현 단계 (Implementation Phases)
1. `go get github.com/NimbleMarkets/ntcharts` 패키지 추가.
2. `internal/tui/model.go`에 차트 초기화 로직 구현.
3. `Update` 메서드에서 데이터 `Push` 연동.
4. `View` 메서드에서 반응형 분기 처리 및 차트 렌더링.
5. 리팩토링: `view_utils.go`의 기존 스파크라인 로직 최적화.

## 5. 검증 계획 (Verification)
- 다양한 터미널 크기(창 크기 조절)에서의 레이아웃 전환 테스트.
- Ollama 서버 부하 발생 시 차트 데이터 갱신 정확도 확인.
- 장시간 실행 시 메모리 누수 여부 점검 (포인트 제한 로직 확인).
