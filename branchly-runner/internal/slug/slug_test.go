package slug

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{
			name:   "long english prompt with stop words",
			prompt: "Add pagination to the products endpoint using cursor",
			// stop words removed: to, the, using
			// first 4: add, pagination, products, endpoint
			want: "branchly/add-pagination-products-endpoint",
		},
		{
			name:   "long portuguese prompt with stop words",
			prompt: "Escrever testes unitários para o serviço de pedidos",
			// stop words removed: para, o, de
			// first 4: escrever, testes, unitários, serviço
			want: "branchly/escrever-testes-unit\u00e1rios-servi\u00e7o",
		},
		{
			name:   "short prompt with fewer than 4 relevant words",
			prompt: "Fix bug",
			want:   "branchly/fix-bug",
		},
		{
			name:   "prompt with special characters and numbers",
			prompt: "Fix auth-service v2!",
			// hyphen and ! replaced by space; no stop words
			// words: fix, auth, service, v2
			want: "branchly/fix-auth-service-v2",
		},
		{
			name:   "result that would exceed 50 characters falls back to 3 words",
			prompt: "Refactoring authentication implementation systematically",
			// 4 words: branchly/refactoring-authentication-implementation-systematically = 65 chars > 50
			// 3 words: branchly/refactoring-authentication-implementation = 50 chars <= 50
			want: "branchly/refactoring-authentication-implementation",
		},
		{
			name:   "all stop words fallback to task",
			prompt: "to the a an",
			want:   "branchly/task",
		},
		{
			name:   "empty prompt fallback to task",
			prompt: "",
			want:   "branchly/task",
		},
		{
			name:   "prompt with only punctuation",
			prompt: "!!! --- ???",
			want:   "branchly/task",
		},
		{
			name:   "english stop words via and with",
			prompt: "Deploy via kubernetes with helm",
			// stop words removed: via, with
			// first 4: deploy, kubernetes, helm
			want: "branchly/deploy-kubernetes-helm",
		},
		{
			name:   "mixed case preserved as lowercase",
			prompt: "Create UserDashboard Component",
			// no stop words; words: create, userdashboard, component
			want: "branchly/create-userdashboard-component",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GenerateSlug(tc.prompt)
			if got != tc.want {
				t.Errorf("GenerateSlug(%q)\n got:  %q\n want: %q", tc.prompt, got, tc.want)
			}
			runes := []rune(got)
			if len(runes) > maxChars {
				t.Errorf("GenerateSlug(%q) = %q has %d chars, exceeds limit of %d", tc.prompt, got, len(runes), maxChars)
			}
		})
	}
}
