<tbody id="tbody">
  {{range $val := .}}
  <tr>
    <td>{{$val.Shorty}}</td>
    <td>{{$val.Url}}</td>
    {{end}}
</tbody>
