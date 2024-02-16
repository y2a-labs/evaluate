// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

type Model struct {
	Name    string
	Checked bool
}

var models = []Model{
	{Name: "openchat/openchat-7b", Checked: true},
	{Name: "undi95/toppy-m-7b", Checked: true},
	{Name: "gryphe/mythomax-l2-13b", Checked: false},
	{Name: "nousresearch/nous-hermes-llama2-13b", Checked: false},
}

func ValidationForm() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<form hx-post=\"/post-script\" class=\"flex flex-col space-y-2\" hx-target=\"#testResults\" hx-swap=\"innerHTML\" enctype=\"multipart/form-data\"><label for=\"test-count\"><span>Test Count</span> <input type=\"number\" name=\"test-count\" class=\"rounded-xl\" placeholder=\"Enter how many tests to run\" value=\"1\"></label> <label for=\"filename\"><span>Filename</span> <input type=\"text\" name=\"filename\" class=\"rounded-xl\" placeholder=\"Enter the yaml file name\" value=\"script.yaml\"></label><div for=\"models\"><span class=\"text-lg\">Models:</span><div class=\"grid lg:grid-cols-2 gap-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, model := range models {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<label for=\"models\"><input type=\"checkbox\" name=\"models\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if model.Checked {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(model.Name))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"> <span>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var2 string
			templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(model.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `components/ValidationForm.templ`, Line: 30, Col: 24}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</span></label>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div><button type=\"submit\" class=\"border border-solid bg-slate-200 rounded-xl px-4 py-2\">Submit</button></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
