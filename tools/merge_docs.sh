#!/bin/bash

# docs フォルダ内のすべてのファイルを 1 つのドキュメントにまとめるツール
# LLM へのコンテキスト提供に最適化：論理的な読み順、アンカーリンク付き目次
# ディレクトリ単位でグループ化し、見出しレベルとリンクを自動調整します

DOCS_DIR="$(dirname "$0")/../docs"
OUTPUT_FILE="$DOCS_DIR/COMPLETE_DOCUMENTATION.md"

# 優先度順序でファイルを処理するためのリスト
# 仕様 → 設計 → ガイド の論理的な読み順
PRIORITY_FILES=(
  # ドキュメント概要
  "README.md"
  # 仕様ドキュメント（基盤）
  "specifications/README.md"
  "specifications/core-specification.md"
  "specifications/meta-protocol.md"
  "specifications/worker-interface.md"
  "specifications/orchestrator-spec.md"
  "specifications/logging-specification.md"
  "specifications/testing-strategy.md"
  # 設計ドキュメント
  "design/README.md"
  "design/architecture.md"
  "design/ide-architecture.md"
  "design/orchestrator-persistence-v2.md"
  "design/chat-autopilot.md"
  "design/task-execution-and-visual-grouping.md"
  "design/data-flow.md"
  "design/implementation-guide.md"
  "design/sandbox-policy.md"
  "task-builder-and-golden-test-design.md"
  # CLIエージェント
  "cli-agents/README.md"
  "cli-agents/codex/version-0.65.0.md"
  # 開発ガイド
  "guides/README.md"
  "guides/testing.md"
  "guides/codex-integration.md"
  "guides/cli-subscription.md"
  "guides/gemini-cli.md"
)

# 除外するファイル
EXCLUDED_FILES=(
  "CLAUDE.md"
  "COMPLETE_DOCUMENTATION.md"
  "GEMINI.md"
  "CURRENT_STATUS.md"
)

# ファイルが除外リストに含まれているかチェック
is_excluded() {
  local file_name="$1"
  for excluded in "${EXCLUDED_FILES[@]}"; do
    if [[ "$file_name" == "$excluded" ]]; then
      return 0
    fi
  done
  return 1
}

# ファイルパスからアンカーを生成
# 例: specifications/core-specification.md → specifications-core-specification
generate_anchor() {
  local file_path="$1"
  # 拡張子を削除
  file_path="${file_path%.md}"
  # / を - に変換
  file_path="${file_path//\//-}"
  echo "$file_path"
}

# ファイルパスから見出しテキストを生成
# 例: core-specification.md → Core Specification
generate_heading() {
  local file_path="$1"
  # ファイル拡張子を削除
  file_path="${file_path%.md}"
  file_path="${file_path%.yaml}"

  # パス区切り（/）をスペースに
  local heading=$(echo "$file_path" | sed 's|/| > |g')

  # ハイフンをスペースに変換
  heading=$(echo "$heading" | sed 's/-/ /g')

  # スネークケースをスペース区切りに変換
  heading=$(echo "$heading" | sed 's/_/ /g')

  # キャメルケースをスペース区切りに（簡易版）
  heading=$(echo "$heading" | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g')

  # 単語の最初の文字を大文字に変換
  heading=$(echo "$heading" | awk '{
    for(i=1;i<=NF;i++) {
      $i = toupper(substr($i,1,1)) substr($i,2)
    }
    print
  }')

  echo "$heading"
}

# 見出しレベルを調整（最大 H4 に制限）
# 各ファイルの最初のH1見出しは削除（ファイル見出しとして別途追加されるため）
# コードブロック内の見出しは処理しない
# 引数: ファイルパス（内容はファイルから直接読み込む）
adjust_heading_levels() {
  local file_path="$1"
  local full_path="$2"
  local max_level=4

  # awk で処理：コードブロック外の最初のH1を削除し、残りの見出しレベルを調整
  # BSD awk 互換のため、!演算子を避ける
  # ファイルから直接読み込むことで大きなファイルでも正しく処理
  awk -v max=$max_level '
    BEGIN { first_h1_removed = 0; in_code_block = 0 }

    # コードブロックの開始/終了を検出
    /^```/ {
      if (in_code_block == 0) {
        in_code_block = 1
      } else {
        in_code_block = 0
      }
      print
      next
    }

    # コードブロック内はそのまま出力
    in_code_block == 1 {
      print
      next
    }

    /^# / {
      # 最初のH1見出しは削除
      if (first_h1_removed == 0) {
        first_h1_removed = 1
        next
      }
    }
    /^#+/ {
      # 見出し行の場合、最初の# の個数を数える
      match($0, /^#+/)
      num_hashes = RLENGTH
      # 1つ増やすが、max を超えないようにする
      new_level = num_hashes + 1
      if (new_level > max) new_level = max
      new_hashes = ""
      for (i = 0; i < new_level; i++) new_hashes = new_hashes "#"
      # 残りの内容を取得
      rest = substr($0, num_hashes + 1)
      print new_hashes rest
      next
    }
    { print }
  ' "$full_path"
}

# リンクを変換（パイプ用）
# [text](path/file.md) → [text](#path-file)
convert_links_pipe() {
  local current_dir="$1"

  # .md ファイルへのリンクをアンカーリンクに変換
  # パスを正規化してアンカー形式に変換
  sed -E '
    # 相対パス ../path/file.md を処理
    s|\[([^\]]+)\]\(\.\./([^)]+)\.md\)|[\1](#\2)|g
    # 相対パス ./path/file.md を処理
    s|\[([^\]]+)\]\(\./([^)]+)\.md\)|[\1](#\2)|g
    # 通常のパス path/file.md を処理
    s|\[([^\]]+)\]\(([^)#:]+)\.md\)|[\1](#\2)|g
  ' | sed -E 's|(#[^)]+)/|\1-|g'
}

# ディレクトリパスから見出しを生成
generate_dir_heading() {
  local dir_path="$1"

  # . の場合は "Overview"
  if [ "$dir_path" = "." ]; then
    echo "Overview"
    return
  fi

  # パス区切り（/）をスペースに
  local heading=$(echo "$dir_path" | sed 's|/| > |g')

  # 単語の最初の文字を大文字に変換
  heading=$(echo "$heading" | awk '{
    for(i=1;i<=NF;i++) {
      $i = toupper(substr($i,1,1)) substr($i,2)
    }
    print
  }')

  echo "$heading"
}

# セクション名を取得
get_section_name() {
  local dir_path="$1"
  case "$dir_path" in
    "specifications") echo "Specifications" ;;
    "design") echo "Design" ;;
    "guides") echo "Guides" ;;
    ".") echo "Overview" ;;
    *) echo "$dir_path" ;;
  esac
}

echo "📚 ドキュメント統合開始..."
echo "" > "$OUTPUT_FILE"

# ヘッダー
{
  echo "# Complete Documentation"
  echo ""
  echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
  echo ""
  echo "This document consolidates all documentation from the docs/ directory for LLM context."
  echo ""
  echo "---"
  echo ""
  echo "## Table of Contents"
  echo ""
} >> "$OUTPUT_FILE"

# 目次を生成（優先度順序に基づく）
echo "## 目次を生成中..."

current_section=""
for rel_path in "${PRIORITY_FILES[@]}"; do
  file="$DOCS_DIR/$rel_path"
  if [ -f "$file" ]; then
    dir_path=$(dirname "$rel_path")
    file_name=$(basename "$rel_path")
    anchor=$(generate_anchor "$rel_path")
    heading=$(generate_heading "$file_name")

    # セクションが変わったら見出しを追加
    section_name=$(get_section_name "$dir_path")
    if [ "$current_section" != "$section_name" ]; then
      current_section="$section_name"
      echo "" >> "$OUTPUT_FILE"
      echo "### $section_name" >> "$OUTPUT_FILE"
      echo "" >> "$OUTPUT_FILE"
    fi

    # 目次エントリを追加
    echo "- [$heading](#$anchor)" >> "$OUTPUT_FILE"
  fi
done

# コンテンツセクション
{
  echo ""
  echo "---"
  echo ""
} >> "$OUTPUT_FILE"

# 優先度順序でファイルを処理
echo ""
echo "## Markdown ファイルを処理中..."

current_section=""
for rel_path in "${PRIORITY_FILES[@]}"; do
  file="$DOCS_DIR/$rel_path"
  if [ -f "$file" ]; then
    dir_path=$(dirname "$rel_path")
    file_name=$(basename "$rel_path")
    anchor=$(generate_anchor "$rel_path")

    # セクションが変わった場合は見出しを出力
    section_name=$(get_section_name "$dir_path")
    if [ "$current_section" != "$section_name" ]; then
      current_section="$section_name"

      {
        echo ""
        echo "# $section_name"
        echo ""
      } >> "$OUTPUT_FILE"
    fi

    # ファイル名から見出しを生成
    file_heading=$(generate_heading "$file_name")

    # 見出しレベルを調整（ファイルから直接読み込み）
    adjusted_content=$(adjust_heading_levels "$rel_path" "$file")

    # リンクを変換
    final_content=$(echo "$adjusted_content" | convert_links_pipe "$dir_path")

    # 出力ファイルに追加（アンカー付き）
    {
      echo "<a id=\"$anchor\"></a>"
      echo ""
      echo "## $file_heading"
      echo ""
      echo "**Source**: \`$rel_path\`"
      echo ""
      echo "$final_content"
      echo ""
    } >> "$OUTPUT_FILE"

    echo "✓ $rel_path"
  fi
done

# 優先度リストにないファイルを処理（将来の拡張用）
echo ""
echo "## 追加ファイルを確認中..."

find "$DOCS_DIR" -type f -name "*.md" | sort | while read file; do
  rel_path="${file#$DOCS_DIR/}"
  file_name=$(basename "$rel_path")

  # 除外ファイルをスキップ
  if is_excluded "$file_name"; then
    continue
  fi

  # 優先度リストに含まれているかチェック
  in_priority=0
  for priority_file in "${PRIORITY_FILES[@]}"; do
    if [ "$rel_path" = "$priority_file" ]; then
      in_priority=1
      break
    fi
  done

  # 優先度リストにないファイルを処理
  if [ $in_priority -eq 0 ]; then
    echo "⚠️ 優先度リスト外: $rel_path"

    dir_path=$(dirname "$rel_path")
    anchor=$(generate_anchor "$rel_path")
    file_heading=$(generate_heading "$file_name")

    # 見出しレベルを調整（ファイルから直接読み込み）
    adjusted_content=$(adjust_heading_levels "$rel_path" "$file")

    # リンクを変換
    final_content=$(echo "$adjusted_content" | convert_links_pipe "$dir_path")

    {
      echo "<a id=\"$anchor\"></a>"
      echo ""
      echo "## $file_heading"
      echo ""
      echo "**Source**: \`$rel_path\`"
      echo ""
      echo "$final_content"
      echo ""
    } >> "$OUTPUT_FILE"
  fi
done

# YAML ファイルを処理
echo ""
echo "## YAML ファイルを処理中..."

yaml_count=$(find "$DOCS_DIR" -type f -name "*.yaml" | wc -l | tr -d ' ')
if [ "$yaml_count" -gt 0 ]; then
  {
    echo ""
    echo "---"
    echo ""
    echo "# YAML Sample Files"
    echo ""
  } >> "$OUTPUT_FILE"

  current_dir=""

  find "$DOCS_DIR" -type f -name "*.yaml" | sort | while read file; do
    rel_path="${file#$DOCS_DIR/}"
    dir_path=$(dirname "$rel_path")
    file_name=$(basename "$rel_path")

    # ディレクトリが変わった場合は見出しを出力
    if [ "$current_dir" != "$dir_path" ]; then
      current_dir="$dir_path"
      dir_heading=$(generate_dir_heading "$dir_path")

      {
        echo ""
        echo "## $dir_heading"
        echo ""
      } >> "$OUTPUT_FILE"
    fi

    file_heading=$(generate_heading "$file_name")
    anchor=$(generate_anchor "$rel_path")

    {
      echo "<a id=\"$anchor\"></a>"
      echo ""
      echo "### $file_heading"
      echo ""
      echo "**Source**: \`$rel_path\`"
      echo ""
      echo '```yaml'
      cat "$file"
      echo '```'
      echo ""
    } >> "$OUTPUT_FILE"

    echo "✓ $rel_path"
  done
fi

# 統計情報
FILE_COUNT=$(find "$DOCS_DIR" -type f \( -name "*.md" -o -name "*.yaml" \) | while read f; do
  fname=$(basename "$f")
  is_excluded "$fname" || echo "$f"
done | wc -l | tr -d ' ')

SIZE_KB=$(du -k "$OUTPUT_FILE" | cut -f1)
LINE_COUNT=$(wc -l < "$OUTPUT_FILE" | tr -d ' ')

echo ""
echo "✅ ドキュメント統合完了!"
echo "   出力ファイル: $OUTPUT_FILE"
echo "   処理ファイル数: $FILE_COUNT"
echo "   ファイルサイズ: ${SIZE_KB} KB"
echo "   行数: ${LINE_COUNT}"
