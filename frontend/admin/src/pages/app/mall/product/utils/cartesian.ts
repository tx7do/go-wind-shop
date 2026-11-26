/**
 * 规格组合笛卡尔积（SKU 明细行生成）。
 *
 * 输入各维度的候选值（如 颜色=[白,红]、尺码=[S,M]），输出全部组合
 * （白S/白M/红S/红M）。维度或值为空时返回空数组——调用方以
 * 「默认规格」单行兜底（无规格商品仍需一个可报价的 SKU）。
 */
export interface ComboCell {
  attributeId: number;
  attributeValueId: number;
}

export function cartesianCombos(dimensions: ComboCell[][]): ComboCell[][] {
  const dims = dimensions.filter((d) => d && d.length > 0);
  if (dims.length === 0) return [];

  let acc: ComboCell[][] = dims[0].map((cell) => [cell]);
  for (let i = 1; i < dims.length; i++) {
    const next: ComboCell[][] = [];
    for (const prefix of acc) {
      for (const cell of dims[i]) {
        next.push([...prefix, cell]);
      }
    }
    acc = next;
  }
  return acc;
}

/** 组合签名：按 attributeId 排序后序列化，作为行的稳定身份键（跨回显/diff 复用）。 */
export function comboKey(combos: ComboCell[]): string {
  return [...combos]
    .sort((a, b) => a.attributeId - b.attributeId)
    .map((c) => `${c.attributeId}:${c.attributeValueId}`)
    .join("|");
}
