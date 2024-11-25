set -eu

PATTERNS=".validate.go _deepcopy.gen.go .gen.json gr.gen.go .pb.go _json.gen.go .pb.gw.go .swagger.json .deepcopy.go"

for p in $PATTERNS; do
    rm -f ./**/**/*"${p}"
    rm -f ./**/*"${p}"
done
