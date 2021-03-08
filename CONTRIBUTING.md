# Communication

Before making a proposal, please spend some time investigating / prototyping.
The devil is in the details, and the feedback from Bazel, linker, and compilers
has proven more convincing than our feeble human voices.

If you find a bug, please create an issue, and if you want to pick up a larger
chunk of work (like exposing a new C++ interface in Go), please do the same for
better visibility.

# Coding style

In the absence of documents, please do stick to the style you find in existing
code. For Go, this is rather canonical, for C/C++, we have added .clang-format
and .clang-tidy configurations to avoid major accidents.

Besides that, please keep in mind that this is not a golf course. Simple, and
possibly repetitive code is favoured over ingenious meta tricks that do little
more than make maintainance and debugging more difficult.

# DCO: Sign your work

Please do sign your work certifying that you have the right to pass it on as an
open-source patch. The rules are pretty simple: if you can certify the contents
of the [DCO](DCO) document contained in this repository, then you just add a
line to every git commit message:

    Signed-off-by: Joe Smith <joe@gmail.com>

using your _real name_. You can also add the sign off via `git commit -s`. Note
that _every_ commit needs to have proper sign-off.
