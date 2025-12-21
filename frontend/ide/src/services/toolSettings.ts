// Go側から取得する型定義
export type ToolOption = {
  id: string;
  name: string;
  description: string;
};

export type ModelOption = {
  id: string;
  name: string;
  group: string;
};

type ToolingBridge = {
  GetToolingConfigJSON: () => Promise<string>;
  SetToolingConfigJSON: (raw: string) => Promise<void>;
  GetAvailableTools: () => Promise<ToolOption[]>;
  GetAvailableModels: () => Promise<ModelOption[]>;
  GetModelsForTool: (toolID: string) => Promise<ModelOption[]>;
  ValidateToolModelCombination: (toolID: string, modelID: string) => Promise<boolean>;
};

declare global {
  interface Window {
    go?: {
      main?: {
        App?: ToolingBridge;
      };
    };
  }
}

function getBridge(): ToolingBridge {
  const bridge = window?.go?.main?.App;
  if (!bridge?.GetToolingConfigJSON || !bridge?.SetToolingConfigJSON) {
    throw new Error("Wails runtime is not available");
  }
  return bridge as ToolingBridge;
}

export async function getToolingConfigJSON(): Promise<string> {
  return await getBridge().GetToolingConfigJSON();
}

export async function setToolingConfigJSON(raw: string): Promise<void> {
  await getBridge().SetToolingConfigJSON(raw);
}

export async function getAvailableTools(): Promise<ToolOption[]> {
  const bridge = getBridge();
  if (!bridge.GetAvailableTools) {
    // フォールバック: Go側にAPIがない場合のデフォルト
    return [
      { id: "claude-code", name: "Claude Code", description: "Anthropic Claude CLI" },
      { id: "gemini-cli", name: "Gemini CLI", description: "Google Gemini CLI" },
      { id: "codex-cli", name: "Codex CLI", description: "OpenAI Codex CLI" },
    ];
  }
  return await bridge.GetAvailableTools();
}

export async function getAvailableModels(): Promise<ModelOption[]> {
  const bridge = getBridge();
  if (!bridge.GetAvailableModels) {
    // フォールバック: Go側にAPIがない場合のデフォルト
    return [{ id: "", name: "Default", group: "Auto" }];
  }
  return await bridge.GetAvailableModels();
}

// 指定されたツールでサポートされるモデル一覧を取得
export async function getModelsForTool(toolID: string): Promise<ModelOption[]> {
  const bridge = getBridge();
  if (!bridge.GetModelsForTool) {
    // フォールバック: 全モデルを返す
    return await getAvailableModels();
  }
  return await bridge.GetModelsForTool(toolID);
}

// ツールとモデルの組み合わせが有効かどうかを検証
export async function validateToolModelCombination(toolID: string, modelID: string): Promise<boolean> {
  const bridge = getBridge();
  if (!bridge.ValidateToolModelCombination) {
    // フォールバック: 常に有効
    return true;
  }
  return await bridge.ValidateToolModelCombination(toolID, modelID);
}
