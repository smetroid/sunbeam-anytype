import React from "react";
import { useState } from "react";
import { Action, ActionPanel, List, Clipboard, showToast, Toast } from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { execFileSync } from "child_process";

interface AnytypeObject {
  cmd: string;
  content: string;
  name: string;
  tags: string;
}

function runCommand(binaryPath: string, args: string[]): string {
  try {
    return execFileSync(binaryPath, args, { encoding: "utf-8" });
  } catch {
    return "[]";
  }
}

export default function SearchCommand() {
  const [searchText, setSearchText] = useState("");

  const { data: objects, isLoading } = usePromise(async (query: string): Promise<AnytypeObject[]> => {
    const binaryPath = `${process.env.HOME}/projects/sunbeam-anytype/sunbeam-anytype`;
    const args = query ? ["--tags", query] : [];

    try {
      const stdout = runCommand(binaryPath, args);
      const parsed = JSON.parse(stdout) as AnytypeObject[];
      return parsed.filter((obj) => obj.cmd && obj.cmd.length > 0);
    } catch {
      return [];
    }
  }, [searchText]);

  return (
    <List isLoading={isLoading} onSearchTextChange={setSearchText} searchBarPlaceholder="Search commands (e.g., cmd, shell)...">
      <List.Section title="Commands">
        {objects?.map((obj) => {
          const code = obj.cmd;
          const title = code.length > 80 ? code.substring(0, 80) + "..." : code;

          return (
            <List.Item
              key={obj.name}
              title={title}
              subtitle={obj.tags}
              actions={
                <ActionPanel>
                  <Action
                    title="Copy Command"
                    onAction={async () => {
                      await Clipboard.copy(code);
                      showToast({
                        style: Toast.Style.Success,
                        title: "Copied to clipboard",
                      });
                    }}
                  />
                  <Action
                    title="Copy Content"
                    onAction={async () => {
                      await Clipboard.copy(obj.content);
                      showToast({
                        style: Toast.Style.Success,
                        title: "Content copied to clipboard",
                      });
                    }}
                  />
                </ActionPanel>
              }
            />
          );
        })}
      </List.Section>
    </List>
  );
}
