import "../../../test/matchMedia";
import { adjustMinValue, adjustMaxValue } from "./index";

jest.mock("chart.js");

describe("common.adjustMinValue", () => {
  test("5", () => {
    expect(adjustMinValue(5)).toEqual(0);
  });

  test("-8", () => {
    expect(adjustMinValue(-8)).toEqual(-9);
  });

  test("-13", () => {
    expect(adjustMinValue(-13)).toEqual(-14);
  });

  test("-20", () => {
    expect(adjustMinValue(-20)).toEqual(-25);
  });

  test("-26", () => {
    expect(adjustMinValue(-26)).toEqual(-30);
  });

  test("-35", () => {
    expect(adjustMinValue(-35)).toEqual(-40);
  });

  test("-50", () => {
    expect(adjustMinValue(-50)).toEqual(-60);
  });

  test("-52", () => {
    expect(adjustMinValue(-52)).toEqual(-60);
  });

  test("-180", () => {
    expect(adjustMinValue(-180)).toEqual(-190);
  });

  test("-210", () => {
    expect(adjustMinValue(-200)).toEqual(-250);
  });

  test("-250", () => {
    expect(adjustMinValue(-250)).toEqual(-300);
  });

  test("-362", () => {
    expect(adjustMinValue(-362)).toEqual(-400);
  });

  test("-1000", () => {
    expect(adjustMinValue(-1000)).toEqual(-1100);
  });

  test("-2363", () => {
    expect(adjustMinValue(-2363)).toEqual(-2400);
  });

  test("-7001", () => {
    expect(adjustMinValue(-7001)).toEqual(-7100);
  });

  test("-10000", () => {
    expect(adjustMinValue(-10000)).toEqual(-11000);
  });

  test("-26526", () => {
    expect(adjustMinValue(-26526)).toEqual(-27000);
  });
});

describe("common.adjustMaxValue", () => {
  test("-5", () => {
    expect(adjustMaxValue(-5)).toEqual(0);
  });

  test("8", () => {
    expect(adjustMaxValue(8)).toEqual(9);
  });

  test("13", () => {
    expect(adjustMaxValue(13)).toEqual(14);
  });

  test("20", () => {
    expect(adjustMaxValue(20)).toEqual(25);
  });

  test("26", () => {
    expect(adjustMaxValue(26)).toEqual(30);
  });

  test("35", () => {
    expect(adjustMaxValue(35)).toEqual(40);
  });

  test("50", () => {
    expect(adjustMaxValue(50)).toEqual(60);
  });

  test("52", () => {
    expect(adjustMaxValue(52)).toEqual(60);
  });

  test("180", () => {
    expect(adjustMaxValue(180)).toEqual(190);
  });

  test("210", () => {
    expect(adjustMaxValue(210)).toEqual(250);
  });

  test("250", () => {
    expect(adjustMaxValue(250)).toEqual(300);
  });

  test("362", () => {
    expect(adjustMaxValue(362)).toEqual(400);
  });

  test("1000", () => {
    expect(adjustMaxValue(1000)).toEqual(1100);
  });

  test("2363", () => {
    expect(adjustMaxValue(2363)).toEqual(2400);
  });

  test("7001", () => {
    expect(adjustMaxValue(7001)).toEqual(7100);
  });

  test("10000", () => {
    expect(adjustMaxValue(10000)).toEqual(11000);
  });

  test("26526", () => {
    expect(adjustMaxValue(26526)).toEqual(27000);
  });
});
