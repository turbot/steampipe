import React from "react";
import SnapshotRenderComplete from "./index.tsx";
import { DashboardContext } from "../../../hooks/useDashboard";
import { render } from "@testing-library/react";
import "@testing-library/jest-dom";

test("return null when should not render snapshot complete div", async () => {
  // ARRANGE
  const { container } = render(
    <DashboardContext.Provider
      value={{ render: { snapshotCompleteDiv: false } }}
    >
      <SnapshotRenderComplete />
    </DashboardContext.Provider>
  );

  // ASSERT
  expect(container).toBeEmptyDOMElement();
});

test("return null when should not render snapshot complete div", async () => {
  // ARRANGE
  render(
    <DashboardContext.Provider
      value={{ render: { snapshotCompleteDiv: true } }}
    >
      <SnapshotRenderComplete />
    </DashboardContext.Provider>
  );

  // ASSERT
  expect(document.querySelector("#snapshot-complete")).toBeTruthy();
});
