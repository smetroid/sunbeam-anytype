import { showToast, Toast } from "@raycast/api";
import { execFileSync } from "child_process";

export default function AddClipboard() {
  const binaryPath = `${process.env.HOME}/projects/sunbeam-anytype/sunbeam-anytype`;

  try {
    execFileSync(binaryPath, ["--clipboard"], { encoding: "utf-8" });
    showToast({
      style: Toast.Style.Success,
      title: "Added to Anytype",
    });
  } catch (error) {
    showToast({
      style: Toast.Style.Failure,
      title: "Failed to add to Anytype",
    });
  }

  return null;
}
