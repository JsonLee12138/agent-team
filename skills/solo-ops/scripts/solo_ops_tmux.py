#!/usr/bin/env python3
"""tmux entrypoint for solo-ops."""

import os


os.environ["SOLO_OPS_BACKEND"] = "tmux"

from solo_ops import main


if __name__ == "__main__":
    main()
