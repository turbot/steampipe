import { icons } from "./heroIcons";

describe("hero icons", () => {
  test("unknown returns null", () => {
    expect(icons["hjdfghskhgskfdjg"]).toEqual(undefined);
  });

  test("existing icon not mapped", () => {
    expect(icons["bell"]).not.toBeNull();
  });

  test("existing namespaced icon not mapped", () => {
    expect(icons["heroicons-outline:bell"]).not.toBeNull();
  });

  test("migrated icon mapped", () => {
    expect(icons["search"]).not.toBeNull();
  });

  test("prefixed migrated icon mapped", () => {
    expect(icons["heroicons-solid:search"]).not.toBeNull();
  });
});
