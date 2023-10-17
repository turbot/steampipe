import { AndFilter, Filter, OrFilter } from "../common"; // Replace with the actual module path
import { validateFilter, validateAndFilter, validateOrFilter } from "./";

describe("CheckFilter", () => {
  describe("validateFilter", () => {
    it("should return true for a valid filter with type and key", () => {
      const filter: Filter = {
        type: "resource",
        key: "name",
      };
      expect(validateFilter(filter)).toBe(true);
    });

    it("should return true for a valid filter with type and value", () => {
      const filter: Filter = {
        type: "tag",
        value: "production",
      };
      expect(validateFilter(filter)).toBe(true);
    });

    it("should return false for a filter missing type", () => {
      // @ts-ignore
      const filter: Filter = {
        key: "name",
      };
      expect(validateFilter(filter)).toBe(false);
    });

    it("should return false for a filter missing both key and value", () => {
      const filter: Filter = {
        type: "resource",
      };
      expect(validateFilter(filter)).toBe(false);
    });
  });

  describe("validateOrFilter", () => {
    it("should return true for a valid OR filter with valid filters", () => {
      const orFilter: OrFilter = {
        or: [
          {
            type: "resource",
            value: "*mybucket*",
          },
          {
            type: "tag",
            key: "environment",
            value: "production",
          },
        ],
      };
      expect(validateOrFilter(orFilter)).toBe(true);
    });

    it("should return true for an empty OR filter", () => {
      const orFilter: OrFilter = { or: [] };
      expect(validateOrFilter(orFilter)).toBe(true);
    });

    it("should return false for an OR filter with an invalid filter", () => {
      const orFilter: OrFilter = {
        or: [
          {
            type: "resource",
            key: "name",
          },
          // @ts-ignore
          {
            key: "name",
          },
        ],
      };
      expect(validateOrFilter(orFilter)).toBe(false);
    });

    it("should return false for an OR filter with one invalid and one valid filter", () => {
      const orFilter: OrFilter = {
        or: [
          {
            type: "resource",
            value: "*mybucket*",
          },
          // @ts-ignore
          {
            key: "name", // This filter is invalid because it's missing the 'type' property.
          },
        ],
      };
      expect(validateOrFilter(orFilter)).toBe(false);
    });
  });

  describe("validateAndFilter", () => {
    it("should return true for a valid AND filter with valid filters", () => {
      const andFilter: AndFilter = {
        and: [
          {
            or: [
              {
                type: "resource",
                value: "*mybucket*",
              },
              {
                type: "tag",
                key: "environment",
                value: "production",
              },
            ],
          },
          {
            type: "dimension",
            key: "region",
            value: "us*",
          },
        ],
      };
      expect(validateAndFilter(andFilter)).toBe(true);
    });

    it("should return true for an empty AND filter", () => {
      const andFilter: AndFilter = { and: [] };
      expect(validateAndFilter(andFilter)).toBe(true);
    });

    it("should return false for an AND filter with an invalid filter", () => {
      const andFilter: AndFilter = {
        and: [
          {
            or: [
              {
                type: "resource",
                key: "name",
              },
              // @ts-ignore
              {
                key: "name",
              },
            ],
          },
          {
            type: "dimension",
            key: "region",
            value: "us*",
          },
        ],
      };
      expect(validateAndFilter(andFilter)).toBe(false);
    });
  });
});
