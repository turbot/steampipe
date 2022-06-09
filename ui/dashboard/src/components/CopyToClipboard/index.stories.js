import React from "react";
import CopyToClipboard from "../CopyToClipboard";

export default {
  component: CopyToClipboard,
  title: "Utilities/Copy to Clipboard",
};

export const NoDataOrPrepareDataFunction = () => (
  <CopyToClipboard data={null} onPrepareData={null} />
);

export const DataPassedIn = () => <CopyToClipboard data="Copy me!" />;

export const KitchenSink = () => (
  <CopyToClipboard
    onPrepareData={async () =>
      await new Promise((resolve) => setTimeout(() => resolve("Copy me!"), 500))
    }
  />
);
