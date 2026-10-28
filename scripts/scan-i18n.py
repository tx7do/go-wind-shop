# -*- coding: utf-8 -*-
"""i18n 键完备性扫描:源码 t() 字面量 + 动态字符串键(labelKey/meta.title 等) vs locales JSON 键树。"""
import json, io, os, re, collections

ROOT = r"D:\GoProject\go-wind-shop\frontend"
EXCLUDE_DIRS = ("node_modules", ".nuxt", ".output", "generated", "dist")


def merge(dst, src):
    for k, v in src.items():
        if isinstance(v, dict) and isinstance(dst.get(k), dict):
            merge(dst[k], v)
        else:
            dst[k] = v


def load_tree(pairs):
    tree = {}
    for ns, obj in pairs:
        node = tree
        for part in ns:
            node = node.setdefault(part, {})
        merge(node, obj)
    return tree


def flatten(tree, prefix=""):
    out = set()
    for k, v in tree.items():
        key = f"{prefix}.{k}" if prefix else k
        out |= flatten(v, key) if isinstance(v, dict) else {key}
    return out


def with_parents(keys):
    out = set(keys)
    for k in keys:
        parts = k.split(".")
        for i in range(1, len(parts)):
            out.add(".".join(parts[:i]))
    return out


def read_src(src_root):
    for dirpath, dirnames, filenames in os.walk(src_root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDE_DIRS]
        for fn in filenames:
            if fn.endswith((".vue", ".ts")):
                p = os.path.join(dirpath, fn)
                try:
                    yield p, io.open(p, encoding="utf-8").read()
                except Exception:
                    pass


def report_missing(files, refs, tag):
    print(f"\n===== {tag} =====")
    avail = {lang: with_parents(flatten(files[lang])) for lang in files}
    missing = {}
    for key, where in refs.items():
        bad = [lang for lang in files if key not in avail[lang]]
        if bad:
            missing[key] = sorted(where)[:2]
    print(f"refs={len(refs)} missing={len(missing)}")
    for key, where in sorted(missing.items()):
        print(f"  {key}   <- {where[0]}")
    return sorted(missing.keys())


def scan_app():
    loc_dir = os.path.join(ROOT, "app", "locales")
    files = {}
    for lang in ["zh-CN", "en-US"]:
        pairs = []
        for fn in os.listdir(os.path.join(loc_dir, lang)):
            if fn.endswith(".json"):
                obj = json.load(io.open(os.path.join(loc_dir, lang, fn), encoding="utf-8"))
                pairs.append(([fn[:-5]], obj))
        files[lang] = load_tree(pairs)

    refs = {}
    tpat = re.compile(r"""[^a-zA-Z_$]t\(\s*['"]([a-zA-Z][\w\.\-]*[a-zA-Z0-9])['"]""")
    for p, s in read_src(os.path.join(ROOT, "app")):
        rel = os.path.relpath(p, os.path.join(ROOT, "app"))
        for m in tpat.finditer(s):
            refs.setdefault(m.group(1), set()).add(rel)
    top_ns = set(files["zh-CN"].keys())
    spat = re.compile(r"""['"]([a-zA-Z][\w\-]*(?:\.[a-zA-Z][\w\-]+)+)['"]""")
    for p, s in read_src(os.path.join(ROOT, "app")):
        rel = os.path.relpath(p, os.path.join(ROOT, "app"))
        for m in spat.finditer(s):
            key = m.group(1)
            if key.split(".")[0] in top_ns:
                refs.setdefault(key, set()).add(rel)

    return report_missing(files, refs, "APP"), files


def scan_admin():
    loc_dir = os.path.join(ROOT, "admin", "src", "locales")
    files = {}
    for lang in ["zh-CN", "en-US"]:
        pairs = []
        base = os.path.join(loc_dir, lang)
        for dirpath, _, filenames in os.walk(base):
            for fn in filenames:
                if not fn.endswith(".json"):
                    continue
                full = os.path.join(dirpath, fn)
                ns = os.path.relpath(full, base).replace("\\", "/")[:-5].split("/")
                obj = json.load(io.open(full, encoding="utf-8"))
                pairs.append((ns, obj))
        files[lang] = load_tree(pairs)

    refs = {}
    tpat = re.compile(r"""[\$\s\(\{,:]t\(\s*['"]([a-zA-Z][\w\.\-]*[a-zA-Z0-9])['"]""")
    for p, s in read_src(os.path.join(ROOT, "admin", "src")):
        rel = os.path.relpath(p, os.path.join(ROOT, "admin", "src"))
        for m in tpat.finditer(s):
            refs.setdefault(m.group(1), set()).add(rel)
    top_ns = set(files["zh-CN"].keys())
    spat = re.compile(r"""['"]([a-zA-Z][\w\-]*(?:\.[a-zA-Z][\w\-]+)+)['"]""")
    for p, s in read_src(os.path.join(ROOT, "admin", "src")):
        rel = os.path.relpath(p, os.path.join(ROOT, "admin", "src"))
        for m in spat.finditer(s):
            key = m.group(1)
            if key.split(".")[0] in top_ns:
                refs.setdefault(key, set()).add(rel)

    return report_missing(files, refs, "ADMIN"), files


app_miss, _ = scan_app()
admin_miss, _ = scan_admin()
print("\nTOTAL app=%d admin=%d" % (len(app_miss), len(admin_miss)))
