import type { Meta, StoryObj } from "@storybook/svelte-vite";
import ChatMessage from "./ChatMessage.svelte";

const meta = {
  title: "Chat/ChatMessage",
  component: ChatMessage,
  tags: ["autodocs"],
  argTypes: {
    role: {
      control: { type: "select" },
      options: ["user", "assistant", "system"],
      description: "メッセージの送信者ロール",
    },
    content: {
      control: "text",
      description: "メッセージ内容",
    },
    timestamp: {
      control: "text",
      description: "タイムスタンプ（ISO 8601形式）",
    },
  },
  parameters: {
    layout: "centered",
    backgrounds: {
      default: "dark",
      values: [{ name: "dark", value: "#2E3440" }],
    },
    docs: {
      description: {
        component:
          "チャットメッセージコンポーネント。ユーザー、アシスタント、システムの3種類のロールに対応しています。ターミナル風のログ表示スタイルを採用。",
      },
    },
  },

} satisfies Meta<ChatMessage>;

export default meta;
type Story = StoryObj<typeof meta>;

// VRT用に固定タイムスタンプを使用（動的な値は視覚回帰テストを不安定にする）
const now = new Date('2024-01-15T10:00:00Z').toISOString();

// ユーザーメッセージ
export const UserMessage: Story = {
  args: {
    role: "user",
    content: "ユーザー認証機能を実装してください",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "ユーザーからのメッセージ。Frost色（水色）で表示されます。",
      },
    },
  },
};

// アシスタントメッセージ
export const AssistantMessage: Story = {
  args: {
    role: "assistant",
    content:
      "承知しました。JWT認証を使用したユーザー認証機能を実装します。まず、認証フローを設計しますね。",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story:
          "アシスタント（Antigravity）からのメッセージ。Aurora Green（緑）で表示されます。",
      },
    },
  },
};

// システムメッセージ
export const SystemMessage: Story = {
  args: {
    role: "system",
    content: "タスク「ユーザー認証機能の実装」が作成されました",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story:
          "システムメッセージ。Aurora Purple（紫）で表示され、イタリック体になります。",
      },
    },
  },
};

// 短いメッセージ
export const ShortMessage: Story = {
  args: {
    role: "user",
    content: "OK",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "短いメッセージの表示。",
      },
    },
  },
};

// 長いメッセージ
export const LongMessage: Story = {
  args: {
    role: "assistant",
    content: `認証機能の実装について詳しく説明します。

まず、以下の手順で進めます：

1. JWT（JSON Web Token）を使用したトークンベース認証の実装
2. ユーザー登録エンドポイントの作成（/api/auth/register）
3. ログインエンドポイントの作成（/api/auth/login）
4. トークン検証ミドルウェアの実装
5. リフレッシュトークンの仕組みの追加
6. パスワードハッシュ化（bcrypt使用）

セキュリティ面では以下に注意します：
- CSRFトークンの検証
- レート制限の実装
- ブルートフォース攻撃対策

実装を開始してもよろしいでしょうか？`,
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "長いメッセージは折り返して表示されます。",
      },
    },
  },
};

// コードを含むメッセージ
export const MessageWithCode: Story = {
  args: {
    role: "assistant",
    content: `認証ミドルウェアのコードサンプルです：

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

このコードをベースに実装を進めます。`,
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "コードを含むメッセージ。等幅フォントで表示されます。",
      },
    },
  },
};

// 日本語と英語混在
export const MixedLanguage: Story = {
  args: {
    role: "user",
    content:
      "The authentication feature should support OAuth2.0 and also handle 日本語のエラーメッセージ for better UX.",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "日本語と英語が混在したメッセージ。",
      },
    },
  },
};

// 過去のタイムスタンプ
export const PastTimestamp: Story = {
  args: {
    role: "user",
    content: "これは過去のメッセージです",
    timestamp: "2024-12-01T09:30:00Z",
  },
  parameters: {
    docs: {
      description: {
        story: "過去のタイムスタンプ表示。",
      },
    },
  },
};

// 絵文字を含むメッセージ
export const WithEmoji: Story = {
  args: {
    role: "assistant",
    content: "タスクが完了しました！ 🎉 次のステップに進みましょう 👍",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "絵文字を含むメッセージ。",
      },
    },
  },
};

// エラーっぽいシステムメッセージ
export const ErrorSystemMessage: Story = {
  args: {
    role: "system",
    content:
      "エラー: タスク「API実装」の実行に失敗しました。詳細はログを確認してください。",
    timestamp: now,
  },
  parameters: {
    docs: {
      description: {
        story: "エラーを示すシステムメッセージ。",
      },
    },
  },
};
