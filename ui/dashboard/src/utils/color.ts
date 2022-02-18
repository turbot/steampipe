import colorConvert from "color-convert";

const minColumn = 16;
const maxColumn = 51;
const minRow = 0;
const maxRow = 5;

interface Color {
  ansi256: number;
  hex: string;
}

interface ColorDictionary {
  [key: number]: boolean;
}

export class ColorGenerator {
  private readonly startingColumn: number;
  private readonly startingRow: number;
  private currentColumn!: number;
  private currentRow!: number;
  private allocatedColorCodes!: ColorDictionary;
  private forbiddenColumns: ColorDictionary;

  constructor(startingColumn: number, startingRow: number) {
    if (startingColumn < minColumn || startingColumn > maxColumn) {
      throw new Error("starting column must be between 16 and 51");
    }
    if (startingRow < minRow || startingRow > maxRow) {
      throw new Error("starting row must be between 0 and 5");
    }

    this.forbiddenColumns = {
      16: true, // red
      17: true, // red
      18: true, // red
      19: true, // red
      20: true, // red
      22: true, // orange
      23: true, // orange
      27: true, // orange
      28: true, // orange
      29: true, // orange
      34: true, // green/orange
      35: true, // green/orange
      36: true, // green/orange
      40: true, // green/orange
      41: true, // green/orange
      42: true, // green/orange
      46: true, // green
      47: true, // green
      48: true, // green
      49: true, // green
    };

    this.startingColumn = startingColumn;
    this.startingRow = startingRow;

    this.reset();
  }

  reset() {
    this.currentColumn = this.startingColumn;
    this.currentRow = this.startingRow;
    this.allocatedColorCodes = {};
  }

  incrementColumn(increment: number) {
    this.currentColumn += increment;
    if (this.currentColumn > maxColumn) {
      // reset and maintain offset
      this.currentColumn -= maxColumn - minColumn + 1;
    }
    while (this.forbiddenColumns[this.currentColumn]) {
      this.currentColumn++;
    }
  }

  incrementRow(increment: number) {
    this.currentRow += increment;
    if (this.currentRow > maxRow) {
      // reset and maintain offset
      this.currentRow -= maxRow;
    }
  }

  colorClashes(color: number) {
    return this.allocatedColorCodes[color];
  }

  currentColor() {
    return this.currentColumn + this.currentRow * 36;
  }

  nextColor(): Color {
    this.incrementColumn(2);
    this.incrementRow(2);

    // does this color clash, or is it forbidden
    let color = this.currentColor();
    const origColor = color;
    while (this.colorClashes(color)) {
      this.incrementColumn(1);
      this.incrementRow(1);
      color = this.currentColor();
      if (color === origColor) {
        // we have tried them all reset and start from the first color
        this.reset();
        return this.nextColor();
      }
    }

    // store this color code
    this.allocatedColorCodes[color] = true;
    return this.toColorObject(color);
  }

  toColorObject(ansi256ColorCode: number): Color {
    return {
      ansi256: ansi256ColorCode,
      hex: `#${colorConvert.ansi256.hex(ansi256ColorCode)}`,
    };
  }
}
