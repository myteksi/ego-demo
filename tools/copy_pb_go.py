#!/usr/bin/env python

import os
import os.path
from shutil import copyfile

print("Current Working Directory " , os.getcwd())
for dirpath, dirnames, filenames in os.walk("./bazel-out"):
    for filename in [f for f in filenames if f.endswith(".pb.go") or f.endswith(".pb.validate.go")]:
        name = os.path.join(dirpath, filename)
        parts = name.split("%/github.com/grab/ego/")
        if len(parts) == 1:
            # print("skipping " + name)
            continue
        dest = parts[1]
        print("copying", name, "-->", dest)
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        copyfile(name, dest)
