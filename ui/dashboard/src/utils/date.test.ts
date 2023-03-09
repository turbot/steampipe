import { timestampForFilename } from "./date";

describe("date utils", () => {
  describe("timestampForFilename", () => {
    test("date correctly formatted", () => {
      const toFormat = new Date(2022, 5, 22, 9, 6, 2);
      expect(timestampForFilename(toFormat)).toEqual("20220622T090602");
    });

    test("number correctly formatted", () => {
      const toFormat = new Date(2022, 5, 22, 9, 6, 2).valueOf();
      expect(timestampForFilename(toFormat)).toEqual("20220622T090602");
    });
  });
});
