/** 平台流量充值类型（与 live-api recharge_type 一致） */
export const RECHARGE_TYPE_LABELS = {
	purchase: '购买',
	gift: '赠送',
	adjust: '调账',
	reduce: '减少',
	initial: '开户',
};

/** 充值表单可选类型（不含开户，开户仅创建商城时自动写入） */
export const RECHARGE_TYPE_FORM_OPTIONS = [
	{ label: '购买', value: 'purchase' },
	{ label: '赠送', value: 'gift' },
	{ label: '调账', value: 'adjust' },
];

export function rechargeTypeLabel(value: unknown) {
	const key = String(value || '').trim();
	return RECHARGE_TYPE_LABELS[key as keyof typeof RECHARGE_TYPE_LABELS] || key || '—';
}
