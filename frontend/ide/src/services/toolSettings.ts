type ToolingBridge = {
  GetToolingConfigJSON: () => Promise<string>;
  SetToolingConfigJSON: (raw: string) => Promise<void>;
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
  return bridge;
}

export async function getToolingConfigJSON(): Promise<string> {
  return await getBridge().GetToolingConfigJSON();
}

export async function setToolingConfigJSON(raw: string): Promise<void> {
  await getBridge().SetToolingConfigJSON(raw);
}
