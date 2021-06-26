load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "69de5c704a05ff37862f7e0f5534d4f479418afc21806c887db544a316f3cb6b",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
    ],
)

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "59d5b42ac315e7eadffa944e86e90c2990110a1c8075f1cd145f487e999d22b3",
    strip_prefix = "rules_docker-0.17.0",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.17.0/rules_docker-v0.17.0.tar.gz"],
)

git_repository(
    name = "distroless",
    commit = "fd0d99e8c54d7d7b2f3dd29f5093d030d192cbbc",
    remote = "https://github.com/GoogleContainerTools/distroless",
    shallow_since = "1582213526 -0500",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

## The below is managed by gazelle.
## If you are changing/adding Go dependencies just run the below command 
## to automatically add the dependencies.
## bazel run //:gazelle -- update -from_file=go.mod

go_repository(
    name = "com_github_coreos_go_iptables",
    importpath = "github.com/coreos/go-iptables",
    sum = "h1:is9qnZMPYjLd8LYqmm/qlE+wwEgJIkTYdhV3rfZo4jk=",
    version = "v0.6.0",
)

go_repository(
    name = "com_github_golang_glog",
    importpath = "github.com/golang/glog",
    sum = "h1:2voWjNECnrZRbfwXxHB1/j8wa6xdKn85B5NzgVL/pTU=",
    version = "v0.0.0-20210429001901-424d2337a529",
)

go_repository(
    name = "com_github_google_gopacket",
    importpath = "github.com/google/gopacket",
    sum = "h1:ves8RnFZPGiFnTS0uPQStjwru6uO6h+nlr9j6fL7kF8=",
    version = "v1.1.19",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    sum = "h1:ObdrDkeb4kJdCP557AjRjq69pTHfNouLtWZG7j9rPN8=",
    version = "v0.0.0-20191011191535-87dc89f01550",
)

go_repository(
    name = "org_golang_x_lint",
    importpath = "golang.org/x/lint",
    sum = "h1:Wh+f8QHJXR411sJR8/vRBTZ7YapZaRvUcLFFJhusH0k=",
    version = "v0.0.0-20200302205851-738671d3881b",
)

go_repository(
    name = "org_golang_x_mod",
    importpath = "golang.org/x/mod",
    sum = "h1:WG0RUwxtNT4qqaXX3DPA8zHFNm/D9xaBpxzHt1WcA/E=",
    version = "v0.1.1-0.20191105210325-c90efee705ee",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:N66aaryRB3Ax92gH0v3hp1QYZ3zWWCCUR/j8Ifh45Ss=",
    version = "v0.0.0-20191028085509-fe3aa8a45271",
)

go_repository(
    name = "org_golang_x_sync",
    importpath = "golang.org/x/sync",
    sum = "h1:8gQV6CLnAEikrhgkHFbMAEhagSSnXWGV915qUMm9mrU=",
    version = "v0.0.0-20190423024810-112230192c58",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:S/FtSvpNLtFBgjTqcKsRpsa6aVsI6iztaz1bQd9BJwE=",
    version = "v0.0.0-20191029155521-f43be2a4598c",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=",
    version = "v0.3.0",
)

go_repository(
    name = "org_golang_x_tools",
    importpath = "golang.org/x/tools",
    sum = "h1:EBZoQjiKKPaLbPrbpssUfuHtwM6KV/vb4U85g/cigFY=",
    version = "v0.0.0-20200130002326-2f3ba24bd6e7",
)

go_repository(
    name = "org_golang_x_xerrors",
    importpath = "golang.org/x/xerrors",
    sum = "h1:/atklqdjdhuosWIl6AIbOeHJjicWYPqR9bpxqxYG2pA=",
    version = "v0.0.0-20191011141410-1b5146add898",
)

go_rules_dependencies()

go_register_toolchains(version = "1.16")

gazelle_dependencies()

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
)

container_pull(
    name = "base",
    digest = "sha256:2f4a6ca7bdf2a532473610a46b2900e94c9c987925bff20603ba087ac6d919f7",
    registry = "index.docker.io",
    repository = "library/ubuntu",
)

load("@distroless//package_manager:package_manager.bzl", "dpkg_list", "dpkg_src", "package_manager_repositories")

# Copied from https://github.com/GoogleCloudPlatform/distroless/blob/master/WORKSPACE
package_manager_repositories()

dpkg_src(
    name = "debian_buster",
    arch = "amd64",
    distro = "buster",
    sha256 = "b044c73a46671536011a26aedd8490dd31140538264ac12f26dc6dd0b4f0fcb8",
    snapshot = "20210601T022916Z",
    url = "https://snapshot.debian.org/archive",
)

dpkg_src(
    name = "debian_buster_security",
    package_prefix = "https://snapshot.debian.org/archive/debian-security/20210601T022916Z/",
    packages_gz_url = "https://snapshot.debian.org/archive/debian-security/20210601T022916Z/dists/buster/updates/main/binary-amd64/Packages.gz",
    sha256 = "95c73e6151604b8087f027efea11ce5b4fac2391d37c743da08499745d985d91",
)

dpkg_list(
    name = "package_bundle",
    packages = [
        "iptables",
        "libip4tc0",
        "libip6tc0",
        "libiptc0",
        "libnetfilter-conntrack3",
        "libnfnetlink0",
        "libpcap0.8",
        "libxtables12",
    ],
    sources = [
        "@debian_buster_security//file:Packages.json",
        "@debian_buster//file:Packages.json",
    ],
)
