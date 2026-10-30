/**
 * 金额显示：后端所有金额字段（SkuPrice.amount、订单/发票/退款的 totalAmount 等）
 * 以「分」为单位、且经 protobuf JSON 序列化成**字符串**下发（uint64→string），
 * 所以这里先 Number() 再除 100，不能直接做字符串运算。
 */

/** 分 → 元（两位小数的数字） */
export function fenToYuan(fen: unknown): number {
  const n = typeof fen === 'string' ? Number(fen) : (fen as number);
  if (!Number.isFinite(n) || n <= 0) return 0;
  return n / 100;
}

/** 分 → 「1234.56」这样的展示串（不带币种符号） */
export function formatFen(fen: unknown, decimals = 2): string {
  return fenToYuan(fen).toFixed(decimals);
}

/** 拆分整数/小数，供价格标签用不同字号排（入参单位：分） */
export function splitYuan(fen: unknown): { int: string; dec: string } {
  const [int, dec] = formatFen(fen).split('.');
  return { int: int ?? '0', dec: dec ? `.${dec}` : '' };
}
