#!/bin/bash

# docs フォルダ内のすべてのファイルを 1 つのドキュメントにまとめるツール
# ディレクトリ単位でグループ化し、見出しレベルとリンクを自動調整します

DOCS_DIR="$(dirname "$0")/../docs"
OUTPUT_FILE="$DOCS_DIR/COMPLETE_DOCUMENTATION.md"

# ファイルパスから見出しテキストを生成
# 例: game-assets/cli-reference.md → CLI リファレンス
generate_heading() {
  local file_path="$1"
  # ファイル拡張子を削除
  file_path="${file_path%.md}"
  file_path="${file_path%.yaml}"

  # パス区切り（/）をスペースに
  local heading=$(echo "$file_path" | sed 's|/| > |g')

  # スネークケースをスペース区切りに変換
  heading=$(echo "$heading" | sed 's/_/ /g')

  # キャメルケースをスペース区切りに（簡易版）
  heading=$(echo "$heading" | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g')

  # 単語の最初の文字を大文字に変換（tr と sed を組み合わせ）
  heading=$(echo "$heading" | awk '{
    for(i=1;i<=NF;i++) {
      $i = toupper(substr($i,1,1)) substr($i,2)
    }
    print
  }')

  echo "$heading"
}

# 見出しレベルを1段階上げる（安全な方法）
# awk を使って見出しの先頭の # の数を数えて増やす
adjust_heading_levels() {
  local file_path="$1"
  local content="$2"

  local is_readme=0
  if [[ "$file_path" == *"README.md" ]]; then
    is_readme=1
  fi

  # README.md の場合、最初のH1見出しのみを削除
  if [ $is_readme -eq 1 ]; then
    content=$(echo "$content" | awk 'NR==1 && /^# / {next} {print}')
  fi

  # awk で見出しレベルを安全に上げる
  content=$(echo "$content" | awk '
    /^#+/ {
      # 見出し行の場合、最初の# の個数を数える
      match($0, /^#+/)
      num_hashes = RLENGTH
      # 1つ増やす
      new_hashes = ""
      for (i = 0; i < num_hashes + 1; i++) new_hashes = new_hashes "#"
      # 残りの内容を取得
      rest = substr($0, num_hashes + 1)
      print new_hashes rest
      next
    }
    { print }
  ')

  echo "$content"
}

# リンクを変換（簡易版）
# [text](path.md) → [text](#anchor)
convert_links() {
  local content="$1"

  # 相対リンク [text](../path/file.md) を [text](#file-name) に変換（簡易）
  # ファイル名からアンカーを生成
  content=$(echo "$content" | sed -E 's|\[([^\]]+)\]\(([^)]+\.md)\)|[\1](#\2)|g')

  echo "$content"
}

# ディレクトリパスから見出しを生成
generate_dir_heading() {
  local dir_path="$1"

  # . の場合は空文字
  if [ "$dir_path" = "." ]; then
    echo "トップレベル"
    return
  fi

  # パス区切り（/）をスペースに
  local heading=$(echo "$dir_path" | sed 's|/| > |g')

  # スネークケースをスペース区切りに変換
  heading=$(echo "$heading" | sed 's/_/ /g')

  # キャメルケースをスペース区切りに（簡易版）
  heading=$(echo "$heading" | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g')

  # 単語の最初の文字を大文字に変換（tr と sed を組み合わせ）
  heading=$(echo "$heading" | awk '{
    for(i=1;i<=NF;i++) {
      $i = toupper(substr($i,1,1)) substr($i,2)
    }
    print
  }')

  echo "$heading"
}

echo "📚 ドキュメント統合開始..."
echo "" > "$OUTPUT_FILE"

# ヘッダー
{
  echo "# 完全なドキュメント"
  echo ""
  echo "生成日: $(date '+%Y-%m-%d %H:%M:%S')"
  echo ""
  echo "このドキュメントは、docs/ ディレクトリ配下のすべてのドキュメントを統合したものです。"
  echo ""
  echo "## 目次"
  echo ""
} >> "$OUTPUT_FILE"

# ディレクトリを集計して目次を作成
echo "## Markdown ファイルを処理中..."

# 一度ファイル一覧を取得して処理
find "$DOCS_DIR" -type f -name "*.md" ! -name "CLAUDE.md" ! -name "COMPLETE_DOCUMENTATION.md" | sort | while read file; do
  rel_path="${file#$DOCS_DIR/}"
  dir_path=$(dirname "$rel_path")

  echo "$dir_path"
done | sort -u | while read dir_path; do
  if [ -n "$dir_path" ]; then
    dir_heading=$(generate_dir_heading "$dir_path")
    echo "- $dir_heading" >> "$OUTPUT_FILE"
  fi
done

# コンテンツを追加
{
  echo ""
  echo "---"
  echo ""
} >> "$OUTPUT_FILE"

# ディレクトリ単位でコンテンツをまとめる
current_dir=""

find "$DOCS_DIR" -type f -name "*.md" ! -name "CLAUDE.md" ! -name "COMPLETE_DOCUMENTATION.md" | sort | while read file; do
  rel_path="${file#$DOCS_DIR/}"
  dir_path=$(dirname "$rel_path")
  file_name=$(basename "$rel_path")

  # ディレクトリが変わった場合は見出しを出力
  if [ "$current_dir" != "$dir_path" ]; then
    current_dir="$dir_path"
    dir_heading=$(generate_dir_heading "$dir_path")

    {
      echo ""
      echo "# $dir_heading"
      echo ""
    } >> "$OUTPUT_FILE"
  fi

  # ファイル名部分のみ（ディレクトリ部分は上で既に出力済み）
  file_heading=$(generate_heading "$file_name")

  # ファイル内容を読み込む
  content=$(cat "$file")

  # 見出しレベルを調整
  content=$(adjust_heading_levels "$rel_path" "$content")

  # リンクを変換
  content=$(convert_links "$content")

  # 出力ファイルに追加
  {
    echo "## $file_heading"
    echo ""
    echo "**ソース**: \`$rel_path\`"
    echo ""
    echo "$content"
    echo ""
  } >> "$OUTPUT_FILE"

  echo "✓ $rel_path"
done

# YAML ファイルを処理
echo ""
echo "## YAML ファイルを処理中..."
{
  echo ""
  echo "---"
  echo ""
  echo "# YAML サンプルファイル"
  echo ""
} >> "$OUTPUT_FILE"

# ディレクトリ単位でYAMLをまとめる
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

  {
    echo "### $file_heading"
    echo ""
    echo "**ソース**: \`$rel_path\`"
    echo ""
    echo '```yaml'
    cat "$file"
    echo '```'
    echo ""
  } >> "$OUTPUT_FILE"

  echo "✓ $rel_path"
done

# 統計情報
FILE_COUNT=$(find "$DOCS_DIR" -type f \( -name "*.md" -o -name "*.yaml" \) ! -name "CLAUDE.md" ! -name "COMPLETE_DOCUMENTATION.md" | wc -l)
SIZE_KB=$(du -k "$OUTPUT_FILE" | cut -f1)

echo ""
echo "✅ ドキュメント統合完了!"
echo "   出力ファイル: $OUTPUT_FILE"
echo "   処理ファイル数: $FILE_COUNT"
echo "   ファイルサイズ: ${SIZE_KB} KB"
